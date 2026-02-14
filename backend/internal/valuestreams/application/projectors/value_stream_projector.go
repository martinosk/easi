package projectors

import (
	"context"
	"encoding/json"
	"log"

	"easi/backend/internal/valuestreams/application/readmodels"
	"easi/backend/internal/valuestreams/domain/events"
	sharedctx "easi/backend/internal/shared/context"
	domain "easi/backend/internal/shared/eventsourcing"
)

type ValueStreamProjector struct {
	readModel *readmodels.ValueStreamReadModel
}

func NewValueStreamProjector(readModel *readmodels.ValueStreamReadModel) *ValueStreamProjector {
	return &ValueStreamProjector{
		readModel: readModel,
	}
}

func (p *ValueStreamProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *ValueStreamProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		"ValueStreamCreated":                p.handleValueStreamCreated,
		"ValueStreamUpdated":                p.handleValueStreamUpdated,
		"ValueStreamDeleted":                p.handleValueStreamDeleted,
		"ValueStreamStageAdded":             p.handleValueStreamStageAdded,
		"ValueStreamStageUpdated":           p.handleValueStreamStageUpdated,
		"ValueStreamStageRemoved":           p.handleValueStreamStageRemoved,
		"ValueStreamStagesReordered":        p.handleValueStreamStagesReordered,
		"ValueStreamStageCapabilityAdded":   p.handleValueStreamStageCapabilityAdded,
		"ValueStreamStageCapabilityRemoved": p.handleValueStreamStageCapabilityRemoved,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

func unmarshalEvent[T any](eventData []byte) (T, error) {
	var event T
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal %T event: %v", event, err)
		return event, err
	}
	return event, nil
}

func projectEvent[T any](eventData []byte, handler func(T) error) error {
	event, err := unmarshalEvent[T](eventData)
	if err != nil {
		return err
	}
	return handler(event)
}

func (p *ValueStreamProjector) handleValueStreamCreated(ctx context.Context, eventData []byte) error {
	return projectEvent(eventData, func(event events.ValueStreamCreated) error {
		return p.readModel.Insert(ctx, readmodels.ValueStreamDTO{
			ID: event.ID, Name: event.Name, Description: event.Description, CreatedAt: event.CreatedAt,
		})
	})
}

func (p *ValueStreamProjector) handleValueStreamUpdated(ctx context.Context, eventData []byte) error {
	return projectEvent(eventData, func(event events.ValueStreamUpdated) error {
		return p.readModel.Update(ctx, event.ID, readmodels.ValueStreamUpdate{
			Name: event.Name, Description: event.Description,
		})
	})
}

func (p *ValueStreamProjector) handleValueStreamDeleted(ctx context.Context, eventData []byte) error {
	return projectEvent(eventData, func(event events.ValueStreamDeleted) error {
		if err := p.readModel.DeleteStagesByValueStreamID(ctx, event.ID); err != nil {
			log.Printf("Failed to delete stages for value stream %s: %v", event.ID, err)
			return err
		}
		return p.readModel.Delete(ctx, event.ID)
	})
}

func (p *ValueStreamProjector) handleValueStreamStageAdded(ctx context.Context, eventData []byte) error {
	return projectEvent(eventData, func(event events.ValueStreamStageAdded) error {
		if err := p.readModel.InsertStage(ctx, readmodels.ValueStreamStageDTO{
			ID: event.StageID, ValueStreamID: event.ID, Name: event.Name, Description: event.Description, Position: event.Position,
		}); err != nil {
			return err
		}
		return p.readModel.AdjustStageCount(ctx, event.ID, 1)
	})
}

func (p *ValueStreamProjector) handleValueStreamStageUpdated(ctx context.Context, eventData []byte) error {
	return projectEvent(eventData, func(event events.ValueStreamStageUpdated) error {
		return p.readModel.UpdateStage(ctx, readmodels.StageUpdate{
			StageID: event.StageID, Name: event.Name, Description: event.Description,
		})
	})
}

func (p *ValueStreamProjector) handleValueStreamStageRemoved(ctx context.Context, eventData []byte) error {
	return projectEvent(eventData, func(event events.ValueStreamStageRemoved) error {
		if err := p.readModel.DeleteStage(ctx, event.StageID); err != nil {
			return err
		}
		return p.readModel.AdjustStageCount(ctx, event.ID, -1)
	})
}

func (p *ValueStreamProjector) handleValueStreamStagesReordered(ctx context.Context, eventData []byte) error {
	event, err := unmarshalEvent[events.ValueStreamStagesReordered](eventData)
	if err != nil {
		return err
	}
	updates := make([]readmodels.StagePositionUpdate, len(event.Positions))
	for i, pos := range event.Positions {
		updates[i] = readmodels.StagePositionUpdate{
			StageID:  pos.StageID,
			Position: pos.Position,
		}
	}
	return p.readModel.UpdateStagePositions(ctx, updates)
}

func (p *ValueStreamProjector) handleValueStreamStageCapabilityAdded(ctx context.Context, eventData []byte) error {
	return p.handleStageCapabilityChange(ctx, eventData, p.readModel.InsertStageCapability)
}

func (p *ValueStreamProjector) handleValueStreamStageCapabilityRemoved(ctx context.Context, eventData []byte) error {
	return p.handleStageCapabilityChange(ctx, eventData, p.readModel.DeleteStageCapability)
}

type stageCapabilityFields struct {
	StageID      string `json:"stageId"`
	CapabilityID string `json:"capabilityId"`
}

func (p *ValueStreamProjector) handleStageCapabilityChange(ctx context.Context, eventData []byte, action func(context.Context, readmodels.StageCapabilityRef) error) error {
	fields, err := unmarshalEvent[stageCapabilityFields](eventData)
	if err != nil {
		return err
	}
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	return action(ctx, readmodels.StageCapabilityRef{
		TenantID: tenantID.Value(), StageID: fields.StageID, CapabilityID: fields.CapabilityID,
	})
}
