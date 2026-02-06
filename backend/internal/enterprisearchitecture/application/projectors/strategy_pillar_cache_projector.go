package projectors

import (
	"context"
	"encoding/json"
	"log"

	"easi/backend/internal/enterprisearchitecture/application/readmodels"
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
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *StrategyPillarCacheProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		"MetaModelConfigurationCreated": p.handleConfigurationCreated,
		"StrategyPillarAdded":           p.handlePillarAdded,
		"StrategyPillarUpdated":         p.handlePillarUpdated,
		"StrategyPillarRemoved":         p.handlePillarRemoved,
		"PillarFitConfigurationUpdated": p.handleFitConfigurationUpdated,
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
		log.Printf("Failed to unmarshal MetaModelConfigurationCreated event: %v", err)
		return err
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
			return err
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
		log.Printf("Failed to unmarshal StrategyPillarAdded event: %v", err)
		return err
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

	return p.readModel.Insert(ctx, dto)
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
		log.Printf("Failed to unmarshal StrategyPillarRemoved event: %v", err)
		return err
	}

	return p.readModel.Delete(ctx, event.PillarID)
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
		log.Printf("Failed to unmarshal pillar event: %v", err)
		return err
	}

	existing, err := p.readModel.GetActivePillar(ctx, event.PillarID)
	if err != nil {
		log.Printf("Failed to get existing pillar %s: %v", event.PillarID, err)
		return err
	}

	if existing == nil {
		existing = &readmodels.StrategyPillarCacheDTO{
			ID:       event.PillarID,
			TenantID: event.TenantID,
			Active:   true,
		}
	}

	mutate(event, existing)
	return p.readModel.Insert(ctx, *existing)
}
