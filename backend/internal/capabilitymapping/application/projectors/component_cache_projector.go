package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	archPL "easi/backend/internal/architecturemodeling/publishedlanguage"
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
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *ComponentCacheProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case archPL.ApplicationComponentCreated:
		return p.handleComponentCreated(ctx, eventData)
	case archPL.ApplicationComponentUpdated:
		return p.handleComponentUpdated(ctx, eventData)
	case archPL.ApplicationComponentDeleted:
		return p.handleComponentDeleted(ctx, eventData)
	}
	return nil
}

type componentCreatedEvent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type componentUpdatedEvent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type componentDeletedEvent struct {
	ID string `json:"id"`
}

func (p *ComponentCacheProjector) handleComponentCreated(ctx context.Context, eventData []byte) error {
	var event componentCreatedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal ApplicationComponentCreated event data: %w", err)
		log.Printf("failed to unmarshal ApplicationComponentCreated event: %v", wrappedErr)
		return wrappedErr
	}
	if err := p.cache.Upsert(ctx, event.ID, event.Name); err != nil {
		return fmt.Errorf("project ApplicationComponentCreated cache upsert for component %s: %w", event.ID, err)
	}
	return nil
}

func (p *ComponentCacheProjector) handleComponentUpdated(ctx context.Context, eventData []byte) error {
	var event componentUpdatedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal ApplicationComponentUpdated event data: %w", err)
		log.Printf("failed to unmarshal ApplicationComponentUpdated event: %v", wrappedErr)
		return wrappedErr
	}
	if err := p.cache.Upsert(ctx, event.ID, event.Name); err != nil {
		return fmt.Errorf("project ApplicationComponentUpdated cache upsert for component %s: %w", event.ID, err)
	}
	return nil
}

func (p *ComponentCacheProjector) handleComponentDeleted(ctx context.Context, eventData []byte) error {
	var event componentDeletedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal ApplicationComponentDeleted event data: %w", err)
		log.Printf("failed to unmarshal ApplicationComponentDeleted event: %v", wrappedErr)
		return wrappedErr
	}
	if err := p.cache.Delete(ctx, event.ID); err != nil {
		return fmt.Errorf("project ApplicationComponentDeleted cache delete for component %s: %w", event.ID, err)
	}
	return nil
}
