package projectors

import (
	"context"
	"encoding/json"
	"fmt"
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
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
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
		wrappedErr := fmt.Errorf("unmarshal StrategyImportanceSet event data: %w", err)
		log.Printf("failed to unmarshal StrategyImportanceSet event: %v", wrappedErr)
		return wrappedErr
	}

	return p.recomputer.RecomputeCapabilityAndDescendants(ctx, ImportanceScope{
		CapabilityID:     event.CapabilityID,
		PillarID:         event.PillarID,
		BusinessDomainID: event.BusinessDomainID,
	})
}

func (p *ImportanceChangeEffectiveProjector) handleStrategyImportanceUpdated(ctx context.Context, eventData []byte) error {
	var event events.StrategyImportanceUpdated
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal StrategyImportanceUpdated event data: %w", err)
		log.Printf("failed to unmarshal StrategyImportanceUpdated event: %v", wrappedErr)
		return wrappedErr
	}

	importance, err := p.importanceReadModel.GetByID(ctx, event.ID)
	if err != nil {
		log.Printf("Failed to get strategy importance %s: %v", event.ID, err)
		return fmt.Errorf("load strategy importance %s for effective recomputation: %w", event.ID, err)
	}
	if importance == nil {
		log.Printf("Strategy importance %s not found for recomputation", event.ID)
		return fmt.Errorf("strategy importance %s not found for recomputation", event.ID)
	}

	return p.recomputer.RecomputeCapabilityAndDescendants(ctx, ImportanceScope{
		CapabilityID:     importance.CapabilityID,
		PillarID:         importance.PillarID,
		BusinessDomainID: importance.BusinessDomainID,
	})
}

func (p *ImportanceChangeEffectiveProjector) handleStrategyImportanceRemoved(ctx context.Context, eventData []byte) error {
	var event events.StrategyImportanceRemoved
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal StrategyImportanceRemoved event data: %w", err)
		log.Printf("failed to unmarshal StrategyImportanceRemoved event: %v", wrappedErr)
		return wrappedErr
	}

	return p.recomputer.RecomputeCapabilityAndDescendants(ctx, ImportanceScope{
		CapabilityID:     event.CapabilityID,
		PillarID:         event.PillarID,
		BusinessDomainID: event.BusinessDomainID,
	})
}
