package projectors

import (
	"context"
	"encoding/json"
	"log"

	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/domain/events"
	domain "easi/backend/internal/shared/eventsourcing"
)

type InternalTeamProjector struct {
	readModel *readmodels.InternalTeamReadModel
}

func NewInternalTeamProjector(readModel *readmodels.InternalTeamReadModel) *InternalTeamProjector {
	return &InternalTeamProjector{
		readModel: readModel,
	}
}

func (p *InternalTeamProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *InternalTeamProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case "InternalTeamCreated":
		return p.projectCreated(ctx, eventData)
	case "InternalTeamUpdated":
		return p.projectUpdated(ctx, eventData)
	case "InternalTeamDeleted":
		return p.projectDeleted(ctx, eventData)
	}
	return nil
}

func (p *InternalTeamProjector) projectCreated(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.InternalTeamCreated](eventData, "InternalTeamCreated")
	if err != nil {
		return err
	}

	return p.readModel.Insert(ctx, readmodels.InternalTeamDTO{
		ID:            event.ID,
		Name:          event.Name,
		Department:    event.Department,
		ContactPerson: event.ContactPerson,
		Notes:         event.Notes,
		CreatedAt:     event.CreatedAt,
	})
}

func (p *InternalTeamProjector) projectUpdated(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.InternalTeamUpdated](eventData, "InternalTeamUpdated")
	if err != nil {
		return err
	}
	return p.readModel.Update(ctx, event.ID, event.Name, event.Department, event.ContactPerson, event.Notes)
}

func (p *InternalTeamProjector) projectDeleted(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.InternalTeamDeleted](eventData, "InternalTeamDeleted")
	if err != nil {
		return err
	}
	return p.readModel.MarkAsDeleted(ctx, event.ID, event.DeletedAt)
}
