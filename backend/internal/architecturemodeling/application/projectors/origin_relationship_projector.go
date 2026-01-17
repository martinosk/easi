package projectors

import (
	"context"
	"encoding/json"
	"log"

	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/domain/events"
	domain "easi/backend/internal/shared/eventsourcing"
)

type OriginRelationshipProjector struct {
	acquiredViaReadModel    *readmodels.AcquiredViaRelationshipReadModel
	purchasedFromReadModel  *readmodels.PurchasedFromRelationshipReadModel
	builtByReadModel        *readmodels.BuiltByRelationshipReadModel
}

func NewOriginRelationshipProjector(
	acquiredViaReadModel *readmodels.AcquiredViaRelationshipReadModel,
	purchasedFromReadModel *readmodels.PurchasedFromRelationshipReadModel,
	builtByReadModel *readmodels.BuiltByRelationshipReadModel,
) *OriginRelationshipProjector {
	return &OriginRelationshipProjector{
		acquiredViaReadModel:   acquiredViaReadModel,
		purchasedFromReadModel: purchasedFromReadModel,
		builtByReadModel:       builtByReadModel,
	}
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
	switch eventType {
	case "AcquiredViaRelationshipCreated":
		return p.projectAcquiredViaCreated(ctx, eventData)
	case "AcquiredViaRelationshipDeleted":
		return p.projectAcquiredViaDeleted(ctx, eventData)
	case "PurchasedFromRelationshipCreated":
		return p.projectPurchasedFromCreated(ctx, eventData)
	case "PurchasedFromRelationshipDeleted":
		return p.projectPurchasedFromDeleted(ctx, eventData)
	case "BuiltByRelationshipCreated":
		return p.projectBuiltByCreated(ctx, eventData)
	case "BuiltByRelationshipDeleted":
		return p.projectBuiltByDeleted(ctx, eventData)
	}
	return nil
}

func (p *OriginRelationshipProjector) projectAcquiredViaCreated(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.AcquiredViaRelationshipCreated](eventData, "AcquiredViaRelationshipCreated")
	if err != nil {
		return err
	}

	return p.acquiredViaReadModel.Insert(ctx, readmodels.AcquiredViaRelationshipDTO{
		ID:               event.ID,
		AcquiredEntityID: event.AcquiredEntityID,
		ComponentID:      event.ComponentID,
		Notes:            event.Notes,
		CreatedAt:        event.CreatedAt,
	})
}

func (p *OriginRelationshipProjector) projectAcquiredViaDeleted(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.AcquiredViaRelationshipDeleted](eventData, "AcquiredViaRelationshipDeleted")
	if err != nil {
		return err
	}
	return p.acquiredViaReadModel.MarkAsDeleted(ctx, event.ID, event.DeletedAt)
}

func (p *OriginRelationshipProjector) projectPurchasedFromCreated(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.PurchasedFromRelationshipCreated](eventData, "PurchasedFromRelationshipCreated")
	if err != nil {
		return err
	}

	return p.purchasedFromReadModel.Insert(ctx, readmodels.PurchasedFromRelationshipDTO{
		ID:          event.ID,
		VendorID:    event.VendorID,
		ComponentID: event.ComponentID,
		Notes:       event.Notes,
		CreatedAt:   event.CreatedAt,
	})
}

func (p *OriginRelationshipProjector) projectPurchasedFromDeleted(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.PurchasedFromRelationshipDeleted](eventData, "PurchasedFromRelationshipDeleted")
	if err != nil {
		return err
	}
	return p.purchasedFromReadModel.MarkAsDeleted(ctx, event.ID, event.DeletedAt)
}

func (p *OriginRelationshipProjector) projectBuiltByCreated(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.BuiltByRelationshipCreated](eventData, "BuiltByRelationshipCreated")
	if err != nil {
		return err
	}

	return p.builtByReadModel.Insert(ctx, readmodels.BuiltByRelationshipDTO{
		ID:             event.ID,
		InternalTeamID: event.InternalTeamID,
		ComponentID:    event.ComponentID,
		Notes:          event.Notes,
		CreatedAt:      event.CreatedAt,
	})
}

func (p *OriginRelationshipProjector) projectBuiltByDeleted(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.BuiltByRelationshipDeleted](eventData, "BuiltByRelationshipDeleted")
	if err != nil {
		return err
	}
	return p.builtByReadModel.MarkAsDeleted(ctx, event.ID, event.DeletedAt)
}
