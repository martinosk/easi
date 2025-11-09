package projectors

import (
	"context"
	"encoding/json"
	"log"

	"github.com/easi/backend/internal/architecturemodeling/application/readmodels"
	"github.com/easi/backend/internal/architecturemodeling/domain/events"
)

// ApplicationComponentProjector projects events to read models
type ApplicationComponentProjector struct {
	readModel *readmodels.ApplicationComponentReadModel
}

// NewApplicationComponentProjector creates a new projector
func NewApplicationComponentProjector(readModel *readmodels.ApplicationComponentReadModel) *ApplicationComponentProjector {
	return &ApplicationComponentProjector{
		readModel: readModel,
	}
}

// ProjectEvent projects a domain event to the read model
func (p *ApplicationComponentProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case "ApplicationComponentCreated":
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

	return nil
}
