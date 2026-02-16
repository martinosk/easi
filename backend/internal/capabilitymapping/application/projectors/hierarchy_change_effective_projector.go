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

type HierarchyChangeEffectiveProjector struct {
	recomputer         *EffectiveImportanceRecomputer
	effectiveReadModel *readmodels.EffectiveCapabilityImportanceReadModel
}

func NewHierarchyChangeEffectiveProjector(
	recomputer *EffectiveImportanceRecomputer,
	effectiveReadModel *readmodels.EffectiveCapabilityImportanceReadModel,
) *HierarchyChangeEffectiveProjector {
	return &HierarchyChangeEffectiveProjector{
		recomputer:         recomputer,
		effectiveReadModel: effectiveReadModel,
	}
}

func (p *HierarchyChangeEffectiveProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *HierarchyChangeEffectiveProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		"CapabilityParentChanged": p.handleCapabilityParentChanged,
		"CapabilityDeleted":       p.handleCapabilityDeleted,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

func (p *HierarchyChangeEffectiveProjector) handleCapabilityParentChanged(ctx context.Context, eventData []byte) error {
	var event events.CapabilityParentChanged
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal CapabilityParentChanged event data: %w", err)
		log.Printf("failed to unmarshal CapabilityParentChanged event: %v", wrappedErr)
		return wrappedErr
	}

	seen := make(map[string]bool)

	if err := p.recomputeFromEffectiveImportances(ctx, event.CapabilityID, event.CapabilityID, seen); err != nil {
		return fmt.Errorf("recompute effective importances for changed capability %s: %w", event.CapabilityID, err)
	}

	if event.NewParentID != "" {
		if err := p.recomputeFromEffectiveImportances(ctx, event.NewParentID, event.CapabilityID, seen); err != nil {
			return fmt.Errorf("recompute effective importances for new parent %s and target capability %s: %w", event.NewParentID, event.CapabilityID, err)
		}
	}

	return nil
}

func (p *HierarchyChangeEffectiveProjector) recomputeFromEffectiveImportances(ctx context.Context, sourceCapabilityID, targetCapabilityID string, seen map[string]bool) error {
	importances, err := p.effectiveReadModel.GetByCapability(ctx, sourceCapabilityID)
	if err != nil {
		log.Printf("Failed to get effective importances for capability %s: %v", sourceCapabilityID, err)
		return fmt.Errorf("load effective importances for capability %s: %w", sourceCapabilityID, err)
	}

	for _, ei := range importances {
		key := ei.PillarID + ":" + ei.BusinessDomainID
		if seen[key] {
			continue
		}
		seen[key] = true

		if err := p.recomputer.RecomputeCapabilityAndDescendants(ctx, ImportanceScope{
			CapabilityID:     targetCapabilityID,
			PillarID:         ei.PillarID,
			BusinessDomainID: ei.BusinessDomainID,
		}); err != nil {
			log.Printf("Failed to recompute capability %s for pillar %s domain %s: %v",
				targetCapabilityID, ei.PillarID, ei.BusinessDomainID, err)
			return fmt.Errorf("recompute from effective importance for target %s pillar %s domain %s: %w", targetCapabilityID, ei.PillarID, ei.BusinessDomainID, err)
		}
	}

	return nil
}

func (p *HierarchyChangeEffectiveProjector) handleCapabilityDeleted(ctx context.Context, eventData []byte) error {
	var eventMap map[string]interface{}
	if err := json.Unmarshal(eventData, &eventMap); err != nil {
		wrappedErr := fmt.Errorf("unmarshal CapabilityDeleted event data: %w", err)
		log.Printf("failed to unmarshal CapabilityDeleted event: %v", wrappedErr)
		return wrappedErr
	}

	capabilityID, ok := eventMap["capabilityId"].(string)
	if !ok {
		capabilityID, _ = eventMap["id"].(string)
	}
	if capabilityID == "" {
		log.Printf("CapabilityDeleted event missing capabilityId")
		return fmt.Errorf("CapabilityDeleted payload missing capability identifier")
	}

	if err := p.effectiveReadModel.DeleteByCapability(ctx, capabilityID); err != nil {
		return fmt.Errorf("delete effective importances for deleted capability %s: %w", capabilityID, err)
	}
	return nil
}
