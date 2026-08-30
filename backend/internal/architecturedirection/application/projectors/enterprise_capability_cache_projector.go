package projectors

import (
	"context"
	"encoding/json"
	"fmt"

	"easi/backend/internal/architecturedirection/application/readmodels"
	eaPL "easi/backend/internal/enterprisearchitecture/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type EnterpriseCapabilityCacheStore interface {
	Insert(ctx context.Context, dto readmodels.EnterpriseCapabilityCacheDTO) error
	UpdateDetails(ctx context.Context, dto readmodels.EnterpriseCapabilityCacheDTO) error
	Deactivate(ctx context.Context, id string) error
	UpdateTargetMaturity(ctx context.Context, id string, targetMaturity int) error
}

type EnterpriseCapabilityCacheProjector struct {
	cache EnterpriseCapabilityCacheStore
}

func NewEnterpriseCapabilityCacheProjector(cache EnterpriseCapabilityCacheStore) *EnterpriseCapabilityCacheProjector {
	return &EnterpriseCapabilityCacheProjector{cache: cache}
}

func (p *EnterpriseCapabilityCacheProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	return dispatchReferenceEvent(ctx, event, p.ProjectEvent)
}

type enterpriseCapabilityCacheEvent struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Category       string `json:"category"`
	Active         bool   `json:"active"`
	TargetMaturity int    `json:"targetMaturity"`
}

func (p *EnterpriseCapabilityCacheProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, enterpriseCapabilityCacheEvent) error{
		eaPL.EnterpriseCapabilityCreated:           p.projectCreated,
		eaPL.EnterpriseCapabilityUpdated:           p.projectUpdated,
		eaPL.EnterpriseCapabilityDeleted:           p.projectDeleted,
		eaPL.EnterpriseCapabilityTargetMaturitySet: p.projectTargetMaturitySet,
	}
	handler, ok := handlers[eventType]
	if !ok {
		return nil
	}
	var event enterpriseCapabilityCacheEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("unmarshal %s payload: %w", eventType, err)
	}
	return handler(ctx, event)
}

func (p *EnterpriseCapabilityCacheProjector) projectCreated(ctx context.Context, event enterpriseCapabilityCacheEvent) error {
	if err := p.cache.Insert(ctx, readmodels.EnterpriseCapabilityCacheDTO{ID: event.ID, Name: event.Name, Category: event.Category, Active: event.Active}); err != nil {
		return fmt.Errorf("cache enterprise capability %s: %w", event.ID, err)
	}
	return nil
}

func (p *EnterpriseCapabilityCacheProjector) projectUpdated(ctx context.Context, event enterpriseCapabilityCacheEvent) error {
	if err := p.cache.UpdateDetails(ctx, readmodels.EnterpriseCapabilityCacheDTO{ID: event.ID, Name: event.Name, Category: event.Category}); err != nil {
		return fmt.Errorf("update cached enterprise capability %s: %w", event.ID, err)
	}
	return nil
}

func (p *EnterpriseCapabilityCacheProjector) projectDeleted(ctx context.Context, event enterpriseCapabilityCacheEvent) error {
	if err := p.cache.Deactivate(ctx, event.ID); err != nil {
		return fmt.Errorf("deactivate cached enterprise capability %s: %w", event.ID, err)
	}
	return nil
}

func (p *EnterpriseCapabilityCacheProjector) projectTargetMaturitySet(ctx context.Context, event enterpriseCapabilityCacheEvent) error {
	if err := p.cache.UpdateTargetMaturity(ctx, event.ID, event.TargetMaturity); err != nil {
		return fmt.Errorf("update target maturity of cached enterprise capability %s: %w", event.ID, err)
	}
	return nil
}
