package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	mmPL "easi/backend/internal/metamodel/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type StrategyPillarCacheProjector struct {
	readModel *readmodels.StrategyPillarCacheReadModel
}

func NewStrategyPillarCacheProjector(readModel *readmodels.StrategyPillarCacheReadModel) *StrategyPillarCacheProjector {
	return &StrategyPillarCacheProjector{readModel: readModel}
}

func (p *StrategyPillarCacheProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *StrategyPillarCacheProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		mmPL.MetaModelConfigurationCreated: p.handleConfigurationCreated,
		mmPL.StrategyPillarAdded:           p.handlePillarAdded,
		mmPL.StrategyPillarUpdated:         p.handlePillarUpdated,
		mmPL.StrategyPillarRemoved:         p.handlePillarRemoved,
		mmPL.PillarFitConfigurationUpdated: p.handleFitConfigurationUpdated,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

type configurationCreatedPillar struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	Active            bool   `json:"active"`
	FitScoringEnabled bool   `json:"fitScoringEnabled"`
	FitCriteria       string `json:"fitCriteria"`
	FitType           string `json:"fitType"`
}

type configurationCreatedEvent struct {
	TenantID string                       `json:"tenantId"`
	Pillars  []configurationCreatedPillar `json:"pillars"`
}

func (p *StrategyPillarCacheProjector) handleConfigurationCreated(ctx context.Context, eventData []byte) error {
	var event configurationCreatedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal MetaModelConfigurationCreated event data in enterprise strategy pillar cache projector: %w", err)
		log.Printf("failed to unmarshal MetaModelConfigurationCreated event: %v", wrappedErr)
		return wrappedErr
	}

	for _, pillar := range event.Pillars {
		dto := readmodels.StrategyPillarCacheDTO{
			ID:                pillar.ID,
			TenantID:          event.TenantID,
			Name:              pillar.Name,
			Description:       pillar.Description,
			Active:            pillar.Active,
			FitScoringEnabled: pillar.FitScoringEnabled,
			FitCriteria:       pillar.FitCriteria,
			FitType:           pillar.FitType,
		}
		if err := p.readModel.Insert(ctx, dto); err != nil {
			return fmt.Errorf("project MetaModelConfigurationCreated enterprise strategy pillar cache insert for pillar %s: %w", pillar.ID, err)
		}
	}

	return nil
}

type pillarAddedEvent struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenantId"`
	PillarID    string `json:"pillarId"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (p *StrategyPillarCacheProjector) handlePillarAdded(ctx context.Context, eventData []byte) error {
	var event pillarAddedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal StrategyPillarAdded event data: %w", err)
		log.Printf("failed to unmarshal StrategyPillarAdded event: %v", wrappedErr)
		return wrappedErr
	}

	dto := readmodels.StrategyPillarCacheDTO{
		ID:                event.PillarID,
		TenantID:          event.TenantID,
		Name:              event.Name,
		Description:       event.Description,
		Active:            true,
		FitScoringEnabled: false,
		FitCriteria:       "",
		FitType:           "",
	}

	if err := p.readModel.Insert(ctx, dto); err != nil {
		return fmt.Errorf("project StrategyPillarAdded enterprise strategy pillar cache insert for pillar %s: %w", event.PillarID, err)
	}
	return nil
}

type pillarEvent struct {
	ID                string `json:"id"`
	TenantID          string `json:"tenantId"`
	PillarID          string `json:"pillarId"`
	NewName           string `json:"newName"`
	NewDescription    string `json:"newDescription"`
	FitScoringEnabled bool   `json:"fitScoringEnabled"`
	FitCriteria       string `json:"fitCriteria"`
	FitType           string `json:"fitType"`
}

func (p *StrategyPillarCacheProjector) handlePillarUpdated(ctx context.Context, eventData []byte) error {
	return p.unmarshalAndUpdate(ctx, eventData, func(event pillarEvent, existing *readmodels.StrategyPillarCacheDTO) {
		existing.Name = event.NewName
		existing.Description = event.NewDescription
	})
}

func (p *StrategyPillarCacheProjector) handlePillarRemoved(ctx context.Context, eventData []byte) error {
	var event pillarEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal StrategyPillarRemoved event data: %w", err)
		log.Printf("failed to unmarshal StrategyPillarRemoved event: %v", wrappedErr)
		return wrappedErr
	}
	if err := p.readModel.Delete(ctx, event.PillarID); err != nil {
		return fmt.Errorf("project StrategyPillarRemoved enterprise strategy pillar cache delete for pillar %s: %w", event.PillarID, err)
	}
	return nil
}

func (p *StrategyPillarCacheProjector) handleFitConfigurationUpdated(ctx context.Context, eventData []byte) error {
	return p.unmarshalAndUpdate(ctx, eventData, func(event pillarEvent, existing *readmodels.StrategyPillarCacheDTO) {
		existing.FitScoringEnabled = event.FitScoringEnabled
		existing.FitCriteria = event.FitCriteria
		existing.FitType = event.FitType
	})
}

func (p *StrategyPillarCacheProjector) unmarshalAndUpdate(ctx context.Context, eventData []byte, mutate func(pillarEvent, *readmodels.StrategyPillarCacheDTO)) error {
	var event pillarEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal strategy pillar event data in enterprise cache projector: %w", err)
		log.Printf("failed to unmarshal pillar event: %v", wrappedErr)
		return wrappedErr
	}

	existing, err := p.readModel.GetActivePillar(ctx, event.PillarID)
	if err != nil {
		log.Printf("Failed to get existing pillar %s: %v", event.PillarID, err)
		return fmt.Errorf("load active enterprise strategy pillar %s for cache mutation: %w", event.PillarID, err)
	}

	if existing == nil {
		existing = &readmodels.StrategyPillarCacheDTO{
			ID:       event.PillarID,
			TenantID: event.TenantID,
			Active:   true,
		}
	}

	mutate(event, existing)
	if err := p.readModel.Insert(ctx, *existing); err != nil {
		return fmt.Errorf("upsert enterprise strategy pillar cache entry for pillar %s: %w", event.PillarID, err)
	}
	return nil
}
