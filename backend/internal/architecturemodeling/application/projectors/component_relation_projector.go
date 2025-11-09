package projectors

import (
	"context"
	"encoding/json"
	"log"

	"github.com/easi/backend/internal/architecturemodeling/application/readmodels"
	"github.com/easi/backend/internal/architecturemodeling/domain/events"
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
	}

	return nil
}
