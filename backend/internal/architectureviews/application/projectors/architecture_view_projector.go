package projectors

import (
	"context"
	"encoding/json"
	"log"

	"easi/backend/internal/architectureviews/application/readmodels"
	"easi/backend/internal/architectureviews/domain/events"
)

// ArchitectureViewProjector projects events to read models
type ArchitectureViewProjector struct {
	readModel *readmodels.ArchitectureViewReadModel
}

// NewArchitectureViewProjector creates a new projector
func NewArchitectureViewProjector(readModel *readmodels.ArchitectureViewReadModel) *ArchitectureViewProjector {
	return &ArchitectureViewProjector{
		readModel: readModel,
	}
}

// ProjectEvent projects a domain event to the read model
func (p *ArchitectureViewProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case "ViewCreated":
		var event events.ViewCreated
		if err := json.Unmarshal(eventData, &event); err != nil {
			log.Printf("Failed to unmarshal ViewCreated event: %v", err)
			return err
		}

		dto := readmodels.ArchitectureViewDTO{
			ID:          event.ID,
			Name:        event.Name,
			Description: event.Description,
			CreatedAt:   event.CreatedAt,
			Components:  make([]readmodels.ComponentPositionDTO, 0),
		}

		return p.readModel.InsertView(ctx, dto)

	case "ComponentAddedToView":
		var event events.ComponentAddedToView
		if err := json.Unmarshal(eventData, &event); err != nil {
			log.Printf("Failed to unmarshal ComponentAddedToView event: %v", err)
			return err
		}

		return p.readModel.AddComponent(ctx, event.ViewID, event.ComponentID, event.X, event.Y)

	case "ComponentPositionUpdated":
		var event events.ComponentPositionUpdated
		if err := json.Unmarshal(eventData, &event); err != nil {
			log.Printf("Failed to unmarshal ComponentPositionUpdated event: %v", err)
			return err
		}

		return p.readModel.UpdateComponentPosition(ctx, event.ViewID, event.ComponentID, event.X, event.Y)
	}

	return nil
}
