package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type BusinessDomainNameCacheWriter interface {
	Upsert(ctx context.Context, businessDomainID, name string) error
	Delete(ctx context.Context, businessDomainID string) error
}

type BusinessDomainNameCacheProjector struct {
	readModel BusinessDomainNameCacheWriter
}

func NewBusinessDomainNameCacheProjector(readModel BusinessDomainNameCacheWriter) *BusinessDomainNameCacheProjector {
	return &BusinessDomainNameCacheProjector{readModel: readModel}
}

func (p *BusinessDomainNameCacheProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *BusinessDomainNameCacheProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		cmPL.BusinessDomainCreated: p.handleUpsert(cmPL.BusinessDomainCreated),
		cmPL.BusinessDomainUpdated: p.handleUpsert(cmPL.BusinessDomainUpdated),
		cmPL.BusinessDomainDeleted: p.handleDeleted,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

type businessDomainNameEvent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (p *BusinessDomainNameCacheProjector) handleUpsert(eventName string) func(context.Context, []byte) error {
	return func(ctx context.Context, eventData []byte) error {
		var event businessDomainNameEvent
		if err := json.Unmarshal(eventData, &event); err != nil {
			wrappedErr := fmt.Errorf("unmarshal %s event data in business domain name cache projector: %w", eventName, err)
			log.Printf("failed to unmarshal %s event: %v", eventName, wrappedErr)
			return wrappedErr
		}
		if err := p.readModel.Upsert(ctx, event.ID, event.Name); err != nil {
			return fmt.Errorf("project %s business domain name cache upsert for domain %s: %w", eventName, event.ID, err)
		}
		return nil
	}
}

type businessDomainIdentifiedEvent struct {
	ID string `json:"id"`
}

func (p *BusinessDomainNameCacheProjector) handleDeleted(ctx context.Context, eventData []byte) error {
	var event businessDomainIdentifiedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal %s event data in business domain name cache projector: %w", cmPL.BusinessDomainDeleted, err)
		log.Printf("failed to unmarshal %s event: %v", cmPL.BusinessDomainDeleted, wrappedErr)
		return wrappedErr
	}
	if err := p.readModel.Delete(ctx, event.ID); err != nil {
		return fmt.Errorf("project %s business domain name cache delete for domain %s: %w", cmPL.BusinessDomainDeleted, event.ID, err)
	}
	return nil
}
