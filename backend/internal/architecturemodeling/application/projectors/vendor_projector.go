package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/domain/events"
	domain "easi/backend/internal/shared/eventsourcing"
)

type VendorProjector struct {
	readModel *readmodels.VendorReadModel
}

func NewVendorProjector(readModel *readmodels.VendorReadModel) *VendorProjector {
	return &VendorProjector{
		readModel: readModel,
	}
}

func (p *VendorProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *VendorProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case "VendorCreated":
		return p.projectCreated(ctx, eventData)
	case "VendorUpdated":
		return p.projectUpdated(ctx, eventData)
	case "VendorDeleted":
		return p.projectDeleted(ctx, eventData)
	}
	return nil
}

func (p *VendorProjector) projectCreated(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.VendorCreated](eventData, "VendorCreated")
	if err != nil {
		return fmt.Errorf("decode VendorCreated event payload in projector: %w", err)
	}
	if err := p.readModel.Insert(ctx, readmodels.VendorDTO{
		ID:                    event.ID,
		Name:                  event.Name,
		ImplementationPartner: event.ImplementationPartner,
		Notes:                 event.Notes,
		CreatedAt:             event.CreatedAt,
	}); err != nil {
		return fmt.Errorf("project VendorCreated for vendor %s: %w", event.ID, err)
	}
	return nil
}

func (p *VendorProjector) projectUpdated(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.VendorUpdated](eventData, "VendorUpdated")
	if err != nil {
		return fmt.Errorf("decode VendorUpdated event payload in projector: %w", err)
	}
	if err := p.readModel.Update(ctx, readmodels.VendorUpdate{
		ID: event.ID, Name: event.Name,
		ImplementationPartner: event.ImplementationPartner, Notes: event.Notes,
	}); err != nil {
		return fmt.Errorf("project VendorUpdated for vendor %s: %w", event.ID, err)
	}
	return nil
}

func (p *VendorProjector) projectDeleted(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.VendorDeleted](eventData, "VendorDeleted")
	if err != nil {
		return fmt.Errorf("decode VendorDeleted event payload in projector: %w", err)
	}
	if err := p.readModel.MarkAsDeleted(ctx, event.ID, event.DeletedAt); err != nil {
		return fmt.Errorf("project VendorDeleted for vendor %s: %w", event.ID, err)
	}
	return nil
}
