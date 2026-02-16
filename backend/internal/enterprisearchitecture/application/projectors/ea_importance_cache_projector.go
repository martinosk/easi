package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	domain "easi/backend/internal/shared/eventsourcing"
)

type ImportanceCacheWriter interface {
	Upsert(ctx context.Context, entry readmodels.ImportanceEntry) error
}

type EAImportanceCacheProjector struct {
	readModel ImportanceCacheWriter
}

func NewEAImportanceCacheProjector(readModel ImportanceCacheWriter) *EAImportanceCacheProjector {
	return &EAImportanceCacheProjector{readModel: readModel}
}

func (p *EAImportanceCacheProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *EAImportanceCacheProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		cmPL.EffectiveImportanceRecalculated: p.handleEffectiveImportanceRecalculated,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

type effectiveImportanceRecalculatedEvent struct {
	CapabilityID     string `json:"capabilityId"`
	BusinessDomainID string `json:"businessDomainId"`
	PillarID         string `json:"pillarId"`
	Importance       int    `json:"importance"`
}

func (p *EAImportanceCacheProjector) handleEffectiveImportanceRecalculated(ctx context.Context, eventData []byte) error {
	var event effectiveImportanceRecalculatedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal EffectiveImportanceRecalculated event data: %w", err)
		log.Printf("failed to unmarshal EffectiveImportanceRecalculated event: %v", wrappedErr)
		return wrappedErr
	}
	if err := p.readModel.Upsert(ctx, readmodels.ImportanceEntry{
		CapabilityID:        event.CapabilityID,
		BusinessDomainID:    event.BusinessDomainID,
		PillarID:            event.PillarID,
		EffectiveImportance: event.Importance,
	}); err != nil {
		return fmt.Errorf("project EffectiveImportanceRecalculated cache upsert for capability %s pillar %s domain %s: %w", event.CapabilityID, event.PillarID, event.BusinessDomainID, err)
	}
	return nil
}
