package projectors

import (
	"context"
	"encoding/json"
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
		log.Printf("Failed to marshal event data: %v", err)
		return err
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
		log.Printf("Failed to unmarshal CapabilityParentChanged event: %v", err)
		return err
	}

	seen := make(map[string]bool)

	if err := p.recomputeFromEffectiveImportances(ctx, event.CapabilityID, event.CapabilityID, seen); err != nil {
		return err
	}

	if event.NewParentID != "" {
		if err := p.recomputeFromEffectiveImportances(ctx, event.NewParentID, event.CapabilityID, seen); err != nil {
			return err
		}
	}

	return nil
}

func (p *HierarchyChangeEffectiveProjector) recomputeFromEffectiveImportances(ctx context.Context, sourceCapabilityID, targetCapabilityID string, seen map[string]bool) error {
	importances, err := p.effectiveReadModel.GetByCapability(ctx, sourceCapabilityID)
	if err != nil {
		log.Printf("Failed to get effective importances for capability %s: %v", sourceCapabilityID, err)
		return err
	}

	for _, ei := range importances {
		key := ei.PillarID + ":" + ei.BusinessDomainID
		if seen[key] {
			continue
		}
		seen[key] = true

		if err := p.recomputer.RecomputeCapabilityAndDescendants(ctx, targetCapabilityID, ei.PillarID, ei.BusinessDomainID); err != nil {
			log.Printf("Failed to recompute capability %s for pillar %s domain %s: %v",
				targetCapabilityID, ei.PillarID, ei.BusinessDomainID, err)
		}
	}

	return nil
}

func (p *HierarchyChangeEffectiveProjector) handleCapabilityDeleted(ctx context.Context, eventData []byte) error {
	var eventMap map[string]interface{}
	if err := json.Unmarshal(eventData, &eventMap); err != nil {
		log.Printf("Failed to unmarshal CapabilityDeleted event: %v", err)
		return err
	}

	capabilityID, ok := eventMap["capabilityId"].(string)
	if !ok {
		capabilityID, _ = eventMap["id"].(string)
	}
	if capabilityID == "" {
		log.Printf("CapabilityDeleted event missing capabilityId")
		return nil
	}

	return p.effectiveReadModel.DeleteByCapability(ctx, capabilityID)
}
