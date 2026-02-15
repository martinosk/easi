package projectors

import (
	"context"
	"encoding/json"
	"log"

	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	domain "easi/backend/internal/shared/eventsourcing"
)

type EAImportanceCacheProjector struct {
	readModel *readmodels.EAImportanceCacheReadModel
}

func NewEAImportanceCacheProjector(readModel *readmodels.EAImportanceCacheReadModel) *EAImportanceCacheProjector {
	return &EAImportanceCacheProjector{readModel: readModel}
}

func (p *EAImportanceCacheProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
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
		log.Printf("Failed to unmarshal EffectiveImportanceRecalculated event: %v", err)
		return err
	}

	return p.readModel.Upsert(ctx, readmodels.ImportanceEntry{
		CapabilityID:        event.CapabilityID,
		BusinessDomainID:    event.BusinessDomainID,
		PillarID:            event.PillarID,
		EffectiveImportance: event.Importance,
	})
}
