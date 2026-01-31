package projectors

import (
	"context"
	"encoding/json"
	"log"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"
	domain "easi/backend/internal/shared/eventsourcing"
)

type ImportanceChangeEffectiveProjector struct {
	recomputer          *EffectiveImportanceRecomputer
	importanceReadModel *readmodels.StrategyImportanceReadModel
}

func NewImportanceChangeEffectiveProjector(
	recomputer *EffectiveImportanceRecomputer,
	importanceReadModel *readmodels.StrategyImportanceReadModel,
) *ImportanceChangeEffectiveProjector {
	return &ImportanceChangeEffectiveProjector{
		recomputer:          recomputer,
		importanceReadModel: importanceReadModel,
	}
}

func (p *ImportanceChangeEffectiveProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *ImportanceChangeEffectiveProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		"StrategyImportanceSet":     p.handleStrategyImportanceSet,
		"StrategyImportanceUpdated": p.handleStrategyImportanceUpdated,
		"StrategyImportanceRemoved": p.handleStrategyImportanceRemoved,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

func (p *ImportanceChangeEffectiveProjector) handleStrategyImportanceSet(ctx context.Context, eventData []byte) error {
	var event events.StrategyImportanceSet
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal StrategyImportanceSet event: %v", err)
		return err
	}

	return p.recomputer.RecomputeCapabilityAndDescendants(ctx, event.CapabilityID, event.PillarID, event.BusinessDomainID)
}

func (p *ImportanceChangeEffectiveProjector) handleStrategyImportanceUpdated(ctx context.Context, eventData []byte) error {
	var event events.StrategyImportanceUpdated
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal StrategyImportanceUpdated event: %v", err)
		return err
	}

	importance, err := p.importanceReadModel.GetByID(ctx, event.ID)
	if err != nil {
		log.Printf("Failed to get strategy importance %s: %v", event.ID, err)
		return err
	}
	if importance == nil {
		log.Printf("Strategy importance %s not found for recomputation", event.ID)
		return nil
	}

	return p.recomputer.RecomputeCapabilityAndDescendants(ctx, importance.CapabilityID, importance.PillarID, importance.BusinessDomainID)
}

func (p *ImportanceChangeEffectiveProjector) handleStrategyImportanceRemoved(ctx context.Context, eventData []byte) error {
	var event events.StrategyImportanceRemoved
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal StrategyImportanceRemoved event: %v", err)
		return err
	}

	return p.recomputer.RecomputeCapabilityAndDescendants(ctx, event.CapabilityID, event.PillarID, event.BusinessDomainID)
}
