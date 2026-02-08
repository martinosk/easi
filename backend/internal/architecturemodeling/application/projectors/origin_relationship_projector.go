package projectors

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

type eventHandler func(ctx context.Context, eventData []byte) error

type OriginRelationshipProjector struct {
	acquiredViaReadModel   *readmodels.AcquiredViaRelationshipReadModel
	purchasedFromReadModel *readmodels.PurchasedFromRelationshipReadModel
	builtByReadModel       *readmodels.BuiltByRelationshipReadModel
	handlers               map[string]eventHandler
}

func NewOriginRelationshipProjector(
	acquiredViaReadModel *readmodels.AcquiredViaRelationshipReadModel,
	purchasedFromReadModel *readmodels.PurchasedFromRelationshipReadModel,
	builtByReadModel *readmodels.BuiltByRelationshipReadModel,
) *OriginRelationshipProjector {
	p := &OriginRelationshipProjector{
		acquiredViaReadModel:   acquiredViaReadModel,
		purchasedFromReadModel: purchasedFromReadModel,
		builtByReadModel:       builtByReadModel,
	}
	p.handlers = map[string]eventHandler{
		"OriginLinkSet":          p.projectOriginLinkSet,
		"OriginLinkReplaced":     p.projectOriginLinkReplaced,
		"OriginLinkNotesUpdated": p.projectOriginLinkNotesUpdated,
		"OriginLinkCleared":      p.projectOriginLinkCleared,
		"OriginLinkDeleted":      p.projectOriginLinkDeleted,
	}
	return p
}

func (p *OriginRelationshipProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *OriginRelationshipProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handler, exists := p.handlers[eventType]
	if !exists {
		return nil
	}
	return handler(ctx, eventData)
}

func (p *OriginRelationshipProjector) projectOriginLinkSet(ctx context.Context, eventData []byte) error {
	return unmarshalAndUpsert[events.OriginLinkSet](p, ctx, eventData, "OriginLinkSet", func(e *events.OriginLinkSet) upsertParams {
		return upsertParams{e.OriginType, e.ComponentID, e.EntityID, e.Notes, e.LinkedAt}
	})
}

func (p *OriginRelationshipProjector) projectOriginLinkReplaced(ctx context.Context, eventData []byte) error {
	return unmarshalAndUpsert[events.OriginLinkReplaced](p, ctx, eventData, "OriginLinkReplaced", func(e *events.OriginLinkReplaced) upsertParams {
		return upsertParams{e.OriginType, e.ComponentID, e.NewEntityID, e.Notes, e.LinkedAt}
	})
}

func unmarshalAndUpsert[T any](p *OriginRelationshipProjector, ctx context.Context, eventData []byte, name string, extract func(*T) upsertParams) error {
	event, err := unmarshalEvent[T](eventData, name)
	if err != nil {
		return err
	}
	return p.upsertRelationship(ctx, extract(event))
}

func (p *OriginRelationshipProjector) projectOriginLinkNotesUpdated(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.OriginLinkNotesUpdated](eventData, "OriginLinkNotesUpdated")
	if err != nil {
		return err
	}
	return p.updateNotes(ctx, event.OriginType, event.ComponentID, event.NewNotes)
}

type upsertParams struct {
	originType  string
	componentID string
	entityID    string
	notes       string
	linkedAt    time.Time
}

func (p *OriginRelationshipProjector) upsertRelationship(ctx context.Context, params upsertParams) error {
	switch params.originType {
	case valueobjects.OriginTypeAcquiredVia:
		return p.acquiredViaReadModel.Upsert(ctx, readmodels.AcquiredViaRelationshipDTO{
			ID: params.componentID, AcquiredEntityID: params.entityID,
			ComponentID: params.componentID, Notes: params.notes, CreatedAt: params.linkedAt,
		})
	case valueobjects.OriginTypePurchasedFrom:
		return p.purchasedFromReadModel.Upsert(ctx, readmodels.PurchasedFromRelationshipDTO{
			ID: params.componentID, VendorID: params.entityID,
			ComponentID: params.componentID, Notes: params.notes, CreatedAt: params.linkedAt,
		})
	case valueobjects.OriginTypeBuiltBy:
		return p.builtByReadModel.Upsert(ctx, readmodels.BuiltByRelationshipDTO{
			ID: params.componentID, InternalTeamID: params.entityID,
			ComponentID: params.componentID, Notes: params.notes, CreatedAt: params.linkedAt,
		})
	}
	return nil
}

func (p *OriginRelationshipProjector) updateNotes(ctx context.Context, originType, componentID, notes string) error {
	return p.forOriginType(originType,
		func() error { return p.acquiredViaReadModel.UpdateNotesByComponentID(ctx, componentID, notes) },
		func() error { return p.purchasedFromReadModel.UpdateNotesByComponentID(ctx, componentID, notes) },
		func() error { return p.builtByReadModel.UpdateNotesByComponentID(ctx, componentID, notes) },
	)
}

func (p *OriginRelationshipProjector) projectOriginLinkCleared(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.OriginLinkCleared](eventData, "OriginLinkCleared")
	if err != nil {
		return err
	}
	return p.deleteByComponentID(ctx, event.OriginType, event.ComponentID)
}

func (p *OriginRelationshipProjector) projectOriginLinkDeleted(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.OriginLinkDeleted](eventData, "OriginLinkDeleted")
	if err != nil {
		return err
	}
	return p.deleteByComponentID(ctx, event.OriginType, event.ComponentID)
}

func (p *OriginRelationshipProjector) deleteByComponentID(ctx context.Context, originType, componentID string) error {
	return p.forOriginType(originType,
		func() error { return p.acquiredViaReadModel.DeleteByComponentID(ctx, componentID) },
		func() error { return p.purchasedFromReadModel.DeleteByComponentID(ctx, componentID) },
		func() error { return p.builtByReadModel.DeleteByComponentID(ctx, componentID) },
	)
}

func (p *OriginRelationshipProjector) forOriginType(originType string, acquiredVia, purchasedFrom, builtBy func() error) error {
	switch originType {
	case valueobjects.OriginTypeAcquiredVia:
		return acquiredVia()
	case valueobjects.OriginTypePurchasedFrom:
		return purchasedFrom()
	case valueobjects.OriginTypeBuiltBy:
		return builtBy()
	}
	return nil
}
