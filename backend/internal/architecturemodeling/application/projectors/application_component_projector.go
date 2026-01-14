package projectors

import (
	"context"
	"encoding/json"
	"log"

	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/domain/events"
	domain "easi/backend/internal/shared/eventsourcing"
)

type ApplicationComponentProjector struct {
	readModel *readmodels.ApplicationComponentReadModel
}

func NewApplicationComponentProjector(readModel *readmodels.ApplicationComponentReadModel) *ApplicationComponentProjector {
	return &ApplicationComponentProjector{
		readModel: readModel,
	}
}

func (p *ApplicationComponentProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *ApplicationComponentProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case "ApplicationComponentCreated":
		return p.projectCreated(ctx, eventData)
	case "ApplicationComponentUpdated":
		return p.projectUpdated(ctx, eventData)
	case "ApplicationComponentDeleted":
		return p.projectDeleted(ctx, eventData)
	case "ApplicationComponentExpertAdded":
		return p.projectExpertAdded(ctx, eventData)
	case "ApplicationComponentExpertRemoved":
		return p.projectExpertRemoved(ctx, eventData)
	}
	return nil
}

func (p *ApplicationComponentProjector) projectCreated(ctx context.Context, eventData []byte) error {
	var event events.ApplicationComponentCreated
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal ApplicationComponentCreated event: %v", err)
		return err
	}

	dto := readmodels.ApplicationComponentDTO{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
		CreatedAt:   event.CreatedAt,
	}

	return p.readModel.Insert(ctx, dto)
}

func (p *ApplicationComponentProjector) projectUpdated(ctx context.Context, eventData []byte) error {
	var event events.ApplicationComponentUpdated
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal ApplicationComponentUpdated event: %v", err)
		return err
	}

	return p.readModel.Update(ctx, event.ID, event.Name, event.Description)
}

func (p *ApplicationComponentProjector) projectDeleted(ctx context.Context, eventData []byte) error {
	var event events.ApplicationComponentDeleted
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal ApplicationComponentDeleted event: %v", err)
		return err
	}

	return p.readModel.MarkAsDeleted(ctx, event.ID, event.DeletedAt)
}

func (p *ApplicationComponentProjector) projectExpertAdded(ctx context.Context, eventData []byte) error {
	var event events.ApplicationComponentExpertAdded
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal ApplicationComponentExpertAdded event: %v", err)
		return err
	}

	return p.readModel.AddExpert(ctx, readmodels.ExpertInfo{
		ComponentID: event.ComponentID,
		Name:        event.ExpertName,
		Role:        event.ExpertRole,
		Contact:     event.ContactInfo,
		AddedAt:     event.AddedAt,
	})
}

func (p *ApplicationComponentProjector) projectExpertRemoved(ctx context.Context, eventData []byte) error {
	var event events.ApplicationComponentExpertRemoved
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal ApplicationComponentExpertRemoved event: %v", err)
		return err
	}

	return p.readModel.RemoveExpert(ctx, event.ComponentID, event.ExpertName, event.ExpertRole, event.ContactInfo)
}
