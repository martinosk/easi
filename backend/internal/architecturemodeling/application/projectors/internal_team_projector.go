package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/domain/events"
	archPL "easi/backend/internal/architecturemodeling/publishedlanguage"
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
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *InternalTeamProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case archPL.InternalTeamCreated:
		return p.projectCreated(ctx, eventData)
	case archPL.InternalTeamUpdated:
		return p.projectUpdated(ctx, eventData)
	case archPL.InternalTeamDeleted:
		return p.projectDeleted(ctx, eventData)
	}
	return nil
}

func (p *InternalTeamProjector) projectCreated(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.InternalTeamCreated](eventData, "InternalTeamCreated")
	if err != nil {
		return fmt.Errorf("decode InternalTeamCreated event payload in projector: %w", err)
	}
	if err := p.readModel.Insert(ctx, readmodels.InternalTeamDTO{
		ID:            event.ID,
		Name:          event.Name,
		Department:    event.Department,
		ContactPerson: event.ContactPerson,
		Notes:         event.Notes,
		CreatedAt:     event.CreatedAt,
	}); err != nil {
		return fmt.Errorf("project InternalTeamCreated for team %s: %w", event.ID, err)
	}
	return nil
}

func (p *InternalTeamProjector) projectUpdated(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.InternalTeamUpdated](eventData, "InternalTeamUpdated")
	if err != nil {
		return fmt.Errorf("decode InternalTeamUpdated event payload in projector: %w", err)
	}
	if err := p.readModel.Update(ctx, readmodels.InternalTeamUpdate{
		ID: event.ID, Name: event.Name, Department: event.Department,
		ContactPerson: event.ContactPerson, Notes: event.Notes,
	}); err != nil {
		return fmt.Errorf("project InternalTeamUpdated for team %s: %w", event.ID, err)
	}
	return nil
}

func (p *InternalTeamProjector) projectDeleted(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.InternalTeamDeleted](eventData, "InternalTeamDeleted")
	if err != nil {
		return fmt.Errorf("decode InternalTeamDeleted event payload in projector: %w", err)
	}
	if err := p.readModel.MarkAsDeleted(ctx, event.ID, event.DeletedAt); err != nil {
		return fmt.Errorf("project InternalTeamDeleted for team %s: %w", event.ID, err)
	}
	return nil
}
