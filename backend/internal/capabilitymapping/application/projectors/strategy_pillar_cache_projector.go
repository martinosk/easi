package projectors

import (
	"context"
	"encoding/json"
	"log"

	"easi/backend/internal/capabilitymapping/application/readmodels"
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
		"StrategyPillarAdded":   p.handlePillarAdded,
		"StrategyPillarUpdated": p.handlePillarUpdated,
		"StrategyPillarRemoved": p.handlePillarRemoved,
		"PillarFitConfigurationUpdated": p.handleFitConfigurationUpdated,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
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

type pillarUpdatedEvent struct {
	ID             string `json:"id"`
	TenantID       string `json:"tenantId"`
	PillarID       string `json:"pillarId"`
	NewName        string `json:"newName"`
	NewDescription string `json:"newDescription"`
}

func (p *StrategyPillarCacheProjector) handlePillarUpdated(ctx context.Context, eventData []byte) error {
	var event pillarUpdatedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal StrategyPillarUpdated event: %v", err)
		return err
	}

	existing, err := p.readModel.GetActivePillar(ctx, event.PillarID)
	if err != nil {
		log.Printf("Failed to get existing pillar %s: %v", event.PillarID, err)
		return err
	}

	if existing == nil {
		log.Printf("Pillar %s not found in cache, skipping update", event.PillarID)
		return nil
	}

	existing.Name = event.NewName
	existing.Description = event.NewDescription

	return p.readModel.Insert(ctx, *existing)
}

type pillarRemovedEvent struct {
	ID       string `json:"id"`
	TenantID string `json:"tenantId"`
	PillarID string `json:"pillarId"`
}

func (p *StrategyPillarCacheProjector) handlePillarRemoved(ctx context.Context, eventData []byte) error {
	var event pillarRemovedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal StrategyPillarRemoved event: %v", err)
		return err
	}

	return p.readModel.Delete(ctx, event.PillarID)
}

type fitConfigurationUpdatedEvent struct {
	ID                string `json:"id"`
	TenantID          string `json:"tenantId"`
	PillarID          string `json:"pillarId"`
	FitScoringEnabled bool   `json:"fitScoringEnabled"`
	FitCriteria       string `json:"fitCriteria"`
	FitType           string `json:"fitType"`
}

func (p *StrategyPillarCacheProjector) handleFitConfigurationUpdated(ctx context.Context, eventData []byte) error {
	var event fitConfigurationUpdatedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal PillarFitConfigurationUpdated event: %v", err)
		return err
	}

	existing, err := p.readModel.GetActivePillar(ctx, event.PillarID)
	if err != nil {
		log.Printf("Failed to get existing pillar %s: %v", event.PillarID, err)
		return err
	}

	if existing == nil {
		log.Printf("Pillar %s not found in cache, skipping fit config update", event.PillarID)
		return nil
	}

	existing.FitScoringEnabled = event.FitScoringEnabled
	existing.FitCriteria = event.FitCriteria
	existing.FitType = event.FitType

	return p.readModel.Insert(ctx, *existing)
}
