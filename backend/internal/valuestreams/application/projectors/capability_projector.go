package projectors

import (
	"context"
	"encoding/json"
	"log"

	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type CapabilityCacheWriter interface {
	Upsert(ctx context.Context, id, name string) error
	Delete(ctx context.Context, id string) error
}

type CapabilityProjector struct {
	cache CapabilityCacheWriter
}

func NewCapabilityProjector(cache CapabilityCacheWriter) *CapabilityProjector {
	return &CapabilityProjector{cache: cache}
}

func (p *CapabilityProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *CapabilityProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case cmPL.CapabilityCreated:
		return p.handleCapabilityCreated(ctx, eventData)
	case cmPL.CapabilityUpdated:
		return p.handleCapabilityUpdated(ctx, eventData)
	case cmPL.CapabilityDeleted:
		return p.handleCapabilityDeleted(ctx, eventData)
	}
	return nil
}

type capabilityCreatedEvent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type capabilityUpdatedEvent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type capabilityDeletedEvent struct {
	ID string `json:"id"`
}

func (p *CapabilityProjector) handleCapabilityCreated(ctx context.Context, eventData []byte) error {
	var event capabilityCreatedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityCreated event: %v", err)
		return err
	}
	return p.cache.Upsert(ctx, event.ID, event.Name)
}

func (p *CapabilityProjector) handleCapabilityUpdated(ctx context.Context, eventData []byte) error {
	var event capabilityUpdatedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityUpdated event: %v", err)
		return err
	}
	return p.cache.Upsert(ctx, event.ID, event.Name)
}

func (p *CapabilityProjector) handleCapabilityDeleted(ctx context.Context, eventData []byte) error {
	var event capabilityDeletedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityDeleted event: %v", err)
		return err
	}
	return p.cache.Delete(ctx, event.ID)
}
