package projectors

import (
	"context"
	"encoding/json"
	"log"

	"easi/backend/internal/valuestreams/application/readmodels"
	"easi/backend/internal/valuestreams/domain/events"
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
		"ValueStreamCreated": p.handleValueStreamCreated,
		"ValueStreamUpdated": p.handleValueStreamUpdated,
		"ValueStreamDeleted": p.handleValueStreamDeleted,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

func (p *ValueStreamProjector) handleValueStreamCreated(ctx context.Context, eventData []byte) error {
	var event events.ValueStreamCreated
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal ValueStreamCreated event: %v", err)
		return err
	}

	dto := readmodels.ValueStreamDTO{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
		CreatedAt:   event.CreatedAt,
	}
	return p.readModel.Insert(ctx, dto)
}

func (p *ValueStreamProjector) handleValueStreamUpdated(ctx context.Context, eventData []byte) error {
	var event events.ValueStreamUpdated
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal ValueStreamUpdated event: %v", err)
		return err
	}
	return p.readModel.Update(ctx, event.ID, readmodels.ValueStreamUpdate{
		Name:        event.Name,
		Description: event.Description,
	})
}

func (p *ValueStreamProjector) handleValueStreamDeleted(ctx context.Context, eventData []byte) error {
	var event events.ValueStreamDeleted
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal ValueStreamDeleted event: %v", err)
		return err
	}
	return p.readModel.Delete(ctx, event.ID)
}
