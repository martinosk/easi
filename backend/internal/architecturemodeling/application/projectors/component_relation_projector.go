package projectors

import (
	"context"
	"encoding/json"
	"log"

	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/shared/domain"
)

// ComponentRelationProjector projects events to read models
type ComponentRelationProjector struct {
	readModel *readmodels.ComponentRelationReadModel
}

// NewComponentRelationProjector creates a new projector
func NewComponentRelationProjector(readModel *readmodels.ComponentRelationReadModel) *ComponentRelationProjector {
	return &ComponentRelationProjector{
		readModel: readModel,
	}
}

// Handle implements the EventHandler interface for the event bus
func (p *ComponentRelationProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

// ProjectEvent projects a domain event to the read model
func (p *ComponentRelationProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case "ComponentRelationCreated":
		var event events.ComponentRelationCreated
		if err := json.Unmarshal(eventData, &event); err != nil {
			log.Printf("Failed to unmarshal ComponentRelationCreated event: %v", err)
			return err
		}

		dto := readmodels.ComponentRelationDTO{
			ID:                event.ID,
			SourceComponentID: event.SourceComponentID,
			TargetComponentID: event.TargetComponentID,
			RelationType:      event.RelationType,
			Name:              event.Name,
			Description:       event.Description,
			CreatedAt:         event.CreatedAt,
		}

		return p.readModel.Insert(ctx, dto)
	case "ComponentRelationUpdated":
		var event events.ComponentRelationUpdated
		if err := json.Unmarshal(eventData, &event); err != nil {
			log.Printf("Failed to unmarshal ComponentRelationUpdated event: %v", err)
			return err
		}

		return p.readModel.Update(ctx, event.ID, event.Name, event.Description)
	case "ComponentRelationDeleted":
		var event events.ComponentRelationDeleted
		if err := json.Unmarshal(eventData, &event); err != nil {
			log.Printf("Failed to unmarshal ComponentRelationDeleted event: %v", err)
			return err
		}

		return p.readModel.MarkAsDeleted(ctx, event.ID, event.DeletedAt)
	}

	return nil
}
