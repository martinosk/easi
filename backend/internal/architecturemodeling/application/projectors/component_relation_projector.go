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
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *ComponentRelationProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case archPL.ComponentRelationCreated:
		return p.projectCreated(ctx, eventData)
	case archPL.ComponentRelationUpdated:
		return p.projectUpdated(ctx, eventData)
	case archPL.ComponentRelationDeleted:
		return p.projectDeleted(ctx, eventData)
	}
	return nil
}

func (p *ComponentRelationProjector) projectCreated(ctx context.Context, eventData []byte) error {
	return projectEvent(ctx, eventData, "ComponentRelationCreated", func(ctx context.Context, event *events.ComponentRelationCreated) error {
		return p.readModel.Insert(ctx, readmodels.ComponentRelationDTO{
			ID:                event.ID,
			SourceComponentID: event.SourceComponentID,
			TargetComponentID: event.TargetComponentID,
			RelationType:      event.RelationType,
			Name:              event.Name,
			Description:       event.Description,
			CreatedAt:         event.CreatedAt,
		})
	})
}

func (p *ComponentRelationProjector) projectUpdated(ctx context.Context, eventData []byte) error {
	return projectEvent(ctx, eventData, "ComponentRelationUpdated", func(ctx context.Context, event *events.ComponentRelationUpdated) error {
		return p.readModel.Update(ctx, event.ID, event.Name, event.Description)
	})
}

func (p *ComponentRelationProjector) projectDeleted(ctx context.Context, eventData []byte) error {
	return projectEvent(ctx, eventData, "ComponentRelationDeleted", func(ctx context.Context, event *events.ComponentRelationDeleted) error {
		return p.readModel.MarkAsDeleted(ctx, event.ID, event.DeletedAt)
	})
}
