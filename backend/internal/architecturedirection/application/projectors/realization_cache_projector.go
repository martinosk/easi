package projectors

import (
	"context"
	"encoding/json"
	"fmt"

	"easi/backend/internal/architecturedirection/application/readmodels"
	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type RealizationCacheStore interface {
	SaveDirectRealization(ctx context.Context, dto readmodels.DirectRealizationDTO) error
	RemoveRealization(ctx context.Context, realizationID readmodels.RealizationID) error
	RemoveRealizationsOfCapability(ctx context.Context, capabilityID readmodels.CapabilityID) error
	RemoveRealizationsOfComponent(ctx context.Context, componentID readmodels.ComponentID) error
}

type realizationCacheEvent struct {
	ID           string `json:"id"`
	CapabilityID string `json:"capabilityId"`
	ComponentID  string `json:"componentId"`
}

type realizationCacheHandler func(ctx context.Context, event realizationCacheEvent) error

type RealizationCacheProjector struct {
	cache RealizationCacheStore
}

func NewRealizationCacheProjector(cache RealizationCacheStore) *RealizationCacheProjector {
	return &RealizationCacheProjector{cache: cache}
}

func (p *RealizationCacheProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	return dispatchReferenceEvent(ctx, event, p.ProjectEvent)
}

func (p *RealizationCacheProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]realizationCacheHandler{
		cmPL.SystemLinkedToCapability:    p.projectSystemLinked,
		cmPL.SystemRealizationDeleted:    p.projectRealizationDeleted,
		cmPL.CapabilityDeleted:           p.projectCapabilityDeleted,
		amPL.ApplicationComponentDeleted: p.projectComponentDeleted,
	}
	handler, tracked := handlers[eventType]
	if !tracked {
		return nil
	}
	var event realizationCacheEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("unmarshal %s payload: %w", eventType, err)
	}
	if event.ID == "" {
		return nil
	}
	if err := handler(ctx, event); err != nil {
		return fmt.Errorf("project %s into realization cache: %w", eventType, err)
	}
	return nil
}

func (p *RealizationCacheProjector) projectSystemLinked(ctx context.Context, event realizationCacheEvent) error {
	if event.CapabilityID == "" || event.ComponentID == "" {
		return nil
	}
	return p.cache.SaveDirectRealization(ctx, readmodels.DirectRealizationDTO{
		RealizationID: readmodels.RealizationID(event.ID),
		CapabilityID:  readmodels.CapabilityID(event.CapabilityID),
		ComponentID:   readmodels.ComponentID(event.ComponentID),
	})
}

func (p *RealizationCacheProjector) projectRealizationDeleted(ctx context.Context, event realizationCacheEvent) error {
	return p.cache.RemoveRealization(ctx, readmodels.RealizationID(event.ID))
}

func (p *RealizationCacheProjector) projectCapabilityDeleted(ctx context.Context, event realizationCacheEvent) error {
	return p.cache.RemoveRealizationsOfCapability(ctx, readmodels.CapabilityID(event.ID))
}

func (p *RealizationCacheProjector) projectComponentDeleted(ctx context.Context, event realizationCacheEvent) error {
	return p.cache.RemoveRealizationsOfComponent(ctx, readmodels.ComponentID(event.ID))
}
