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
	case "ComponentOriginsCreated":
		return nil
	case "AcquiredViaRelationshipSet":
		return p.projectAcquiredViaSet(ctx, eventData)
	case "AcquiredViaRelationshipReplaced":
		return p.projectAcquiredViaReplaced(ctx, eventData)
	case "AcquiredViaNotesUpdated":
		return p.projectAcquiredViaNotesUpdated(ctx, eventData)
	case "AcquiredViaRelationshipCleared":
		return p.projectAcquiredViaCleared(ctx, eventData)
	case "PurchasedFromRelationshipSet":
		return p.projectPurchasedFromSet(ctx, eventData)
	case "PurchasedFromRelationshipReplaced":
		return p.projectPurchasedFromReplaced(ctx, eventData)
	case "PurchasedFromNotesUpdated":
		return p.projectPurchasedFromNotesUpdated(ctx, eventData)
	case "PurchasedFromRelationshipCleared":
		return p.projectPurchasedFromCleared(ctx, eventData)
	case "BuiltByRelationshipSet":
		return p.projectBuiltBySet(ctx, eventData)
	case "BuiltByRelationshipReplaced":
		return p.projectBuiltByReplaced(ctx, eventData)
	case "BuiltByNotesUpdated":
		return p.projectBuiltByNotesUpdated(ctx, eventData)
	case "BuiltByRelationshipCleared":
		return p.projectBuiltByCleared(ctx, eventData)
	case "ComponentOriginsDeleted":
		return p.projectComponentOriginsDeleted(ctx, eventData)
	}
	return nil
}

func (p *OriginRelationshipProjector) projectAcquiredViaSet(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.AcquiredViaRelationshipSet](eventData, "AcquiredViaRelationshipSet")
	if err != nil {
		return err
	}

	return p.acquiredViaReadModel.Insert(ctx, readmodels.AcquiredViaRelationshipDTO{
		ID:               event.ComponentID,
		AcquiredEntityID: event.EntityID,
		ComponentID:      event.ComponentID,
		Notes:            event.Notes,
		CreatedAt:        event.LinkedAt,
	})
}

func (p *OriginRelationshipProjector) projectAcquiredViaReplaced(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.AcquiredViaRelationshipReplaced](eventData, "AcquiredViaRelationshipReplaced")
	if err != nil {
		return err
	}

	return p.acquiredViaReadModel.UpdateByComponentID(ctx, readmodels.AcquiredViaRelationshipDTO{
		AcquiredEntityID: event.NewEntityID,
		ComponentID:      event.ComponentID,
		Notes:            event.Notes,
		CreatedAt:        event.LinkedAt,
	})
}

func (p *OriginRelationshipProjector) projectAcquiredViaNotesUpdated(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.AcquiredViaNotesUpdated](eventData, "AcquiredViaNotesUpdated")
	if err != nil {
		return err
	}

	return p.acquiredViaReadModel.UpdateNotesByComponentID(ctx, event.ComponentID, event.NewNotes)
}

func (p *OriginRelationshipProjector) projectAcquiredViaCleared(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.AcquiredViaRelationshipCleared](eventData, "AcquiredViaRelationshipCleared")
	if err != nil {
		return err
	}
	return p.acquiredViaReadModel.DeleteByComponentID(ctx, event.ComponentID)
}

func (p *OriginRelationshipProjector) projectPurchasedFromSet(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.PurchasedFromRelationshipSet](eventData, "PurchasedFromRelationshipSet")
	if err != nil {
		return err
	}

	return p.purchasedFromReadModel.Insert(ctx, readmodels.PurchasedFromRelationshipDTO{
		ID:          event.ComponentID,
		VendorID:    event.VendorID,
		ComponentID: event.ComponentID,
		Notes:       event.Notes,
		CreatedAt:   event.LinkedAt,
	})
}

func (p *OriginRelationshipProjector) projectPurchasedFromReplaced(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.PurchasedFromRelationshipReplaced](eventData, "PurchasedFromRelationshipReplaced")
	if err != nil {
		return err
	}

	return p.purchasedFromReadModel.UpdateByComponentID(ctx, readmodels.PurchasedFromRelationshipDTO{
		VendorID:    event.NewVendorID,
		ComponentID: event.ComponentID,
		Notes:       event.Notes,
		CreatedAt:   event.LinkedAt,
	})
}

func (p *OriginRelationshipProjector) projectPurchasedFromNotesUpdated(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.PurchasedFromNotesUpdated](eventData, "PurchasedFromNotesUpdated")
	if err != nil {
		return err
	}

	return p.purchasedFromReadModel.UpdateNotesByComponentID(ctx, event.ComponentID, event.NewNotes)
}

func (p *OriginRelationshipProjector) projectPurchasedFromCleared(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.PurchasedFromRelationshipCleared](eventData, "PurchasedFromRelationshipCleared")
	if err != nil {
		return err
	}
	return p.purchasedFromReadModel.DeleteByComponentID(ctx, event.ComponentID)
}

func (p *OriginRelationshipProjector) projectBuiltBySet(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.BuiltByRelationshipSet](eventData, "BuiltByRelationshipSet")
	if err != nil {
		return err
	}

	return p.builtByReadModel.Insert(ctx, readmodels.BuiltByRelationshipDTO{
		ID:             event.ComponentID,
		InternalTeamID: event.TeamID,
		ComponentID:    event.ComponentID,
		Notes:          event.Notes,
		CreatedAt:      event.LinkedAt,
	})
}

func (p *OriginRelationshipProjector) projectBuiltByReplaced(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.BuiltByRelationshipReplaced](eventData, "BuiltByRelationshipReplaced")
	if err != nil {
		return err
	}

	return p.builtByReadModel.UpdateByComponentID(ctx, readmodels.BuiltByRelationshipDTO{
		InternalTeamID: event.NewTeamID,
		ComponentID:    event.ComponentID,
		Notes:          event.Notes,
		CreatedAt:      event.LinkedAt,
	})
}

func (p *OriginRelationshipProjector) projectBuiltByNotesUpdated(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.BuiltByNotesUpdated](eventData, "BuiltByNotesUpdated")
	if err != nil {
		return err
	}

	return p.builtByReadModel.UpdateNotesByComponentID(ctx, event.ComponentID, event.NewNotes)
}

func (p *OriginRelationshipProjector) projectBuiltByCleared(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.BuiltByRelationshipCleared](eventData, "BuiltByRelationshipCleared")
	if err != nil {
		return err
	}
	return p.builtByReadModel.DeleteByComponentID(ctx, event.ComponentID)
}

func (p *OriginRelationshipProjector) projectComponentOriginsDeleted(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.ComponentOriginsDeleted](eventData, "ComponentOriginsDeleted")
	if err != nil {
		return err
	}

	if err := p.acquiredViaReadModel.DeleteByComponentID(ctx, event.ComponentID); err != nil {
		return err
	}
	if err := p.purchasedFromReadModel.DeleteByComponentID(ctx, event.ComponentID); err != nil {
		return err
	}
	return p.builtByReadModel.DeleteByComponentID(ctx, event.ComponentID)
}
