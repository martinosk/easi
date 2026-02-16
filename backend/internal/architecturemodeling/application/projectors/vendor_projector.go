package projectors

import (
	"context"
	"encoding/json"
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
		log.Printf("Failed to marshal event data: %v", err)
		return err
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
		return err
	}

	return p.readModel.Insert(ctx, readmodels.VendorDTO{
		ID:                    event.ID,
		Name:                  event.Name,
		ImplementationPartner: event.ImplementationPartner,
		Notes:                 event.Notes,
		CreatedAt:             event.CreatedAt,
	})
}

func (p *VendorProjector) projectUpdated(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.VendorUpdated](eventData, "VendorUpdated")
	if err != nil {
		return err
	}
	return p.readModel.Update(ctx, readmodels.VendorUpdate{
		ID: event.ID, Name: event.Name,
		ImplementationPartner: event.ImplementationPartner, Notes: event.Notes,
	})
}

func (p *VendorProjector) projectDeleted(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.VendorDeleted](eventData, "VendorDeleted")
	if err != nil {
		return err
	}
	return p.readModel.MarkAsDeleted(ctx, event.ID, event.DeletedAt)
}
