package projectors

import (
	"context"
	"encoding/json"
	"log"

	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/domain/events"
	domain "easi/backend/internal/shared/eventsourcing"
)

type AcquiredEntityProjector struct {
	readModel *readmodels.AcquiredEntityReadModel
}

func NewAcquiredEntityProjector(readModel *readmodels.AcquiredEntityReadModel) *AcquiredEntityProjector {
	return &AcquiredEntityProjector{
		readModel: readModel,
	}
}

func (p *AcquiredEntityProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *AcquiredEntityProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case "AcquiredEntityCreated":
		return p.projectCreated(ctx, eventData)
	case "AcquiredEntityUpdated":
		return p.projectUpdated(ctx, eventData)
	case "AcquiredEntityDeleted":
		return p.projectDeleted(ctx, eventData)
	}
	return nil
}

func (p *AcquiredEntityProjector) projectCreated(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.AcquiredEntityCreated](eventData, "AcquiredEntityCreated")
	if err != nil {
		return err
	}

	return p.readModel.Insert(ctx, readmodels.AcquiredEntityDTO{
		ID:                event.ID,
		Name:              event.Name,
		AcquisitionDate:   event.AcquisitionDate,
		IntegrationStatus: event.IntegrationStatus,
		Notes:             event.Notes,
		CreatedAt:         event.CreatedAt,
	})
}

func (p *AcquiredEntityProjector) projectUpdated(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.AcquiredEntityUpdated](eventData, "AcquiredEntityUpdated")
	if err != nil {
		return err
	}
	return p.readModel.Update(ctx, event.ID, event.Name, event.AcquisitionDate, event.IntegrationStatus, event.Notes)
}

func (p *AcquiredEntityProjector) projectDeleted(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.AcquiredEntityDeleted](eventData, "AcquiredEntityDeleted")
	if err != nil {
		return err
	}
	return p.readModel.MarkAsDeleted(ctx, event.ID, event.DeletedAt)
}
