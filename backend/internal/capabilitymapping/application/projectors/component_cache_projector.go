package projectors

import (
	"context"
	"encoding/json"
	"log"

	archEvents "easi/backend/internal/architecturemodeling/domain/events"
	domain "easi/backend/internal/shared/eventsourcing"
)

type ComponentCacheWriter interface {
	Upsert(ctx context.Context, id, name string) error
	Delete(ctx context.Context, id string) error
}

type ComponentCacheProjector struct {
	cache ComponentCacheWriter
}

func NewComponentCacheProjector(cache ComponentCacheWriter) *ComponentCacheProjector {
	return &ComponentCacheProjector{cache: cache}
}

func (p *ComponentCacheProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *ComponentCacheProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case "ApplicationComponentCreated":
		return p.handleComponentCreated(ctx, eventData)
	case "ApplicationComponentUpdated":
		return p.handleComponentUpdated(ctx, eventData)
	case "ApplicationComponentDeleted":
		return p.handleComponentDeleted(ctx, eventData)
	}
	return nil
}

func (p *ComponentCacheProjector) handleComponentCreated(ctx context.Context, eventData []byte) error {
	var event archEvents.ApplicationComponentCreated
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal ApplicationComponentCreated event: %v", err)
		return err
	}
	return p.cache.Upsert(ctx, event.ID, event.Name)
}

func (p *ComponentCacheProjector) handleComponentUpdated(ctx context.Context, eventData []byte) error {
	var event archEvents.ApplicationComponentUpdated
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal ApplicationComponentUpdated event: %v", err)
		return err
	}
	return p.cache.Upsert(ctx, event.ID, event.Name)
}

func (p *ComponentCacheProjector) handleComponentDeleted(ctx context.Context, eventData []byte) error {
	var event archEvents.ApplicationComponentDeleted
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal ApplicationComponentDeleted event: %v", err)
		return err
	}
	return p.cache.Delete(ctx, event.ID)
}
