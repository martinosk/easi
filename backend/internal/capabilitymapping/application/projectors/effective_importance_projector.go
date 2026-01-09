package projectors

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/capabilitymapping/domain/services"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/metamodel"
	domain "easi/backend/internal/shared/eventsourcing"
)

type DomainAssignmentChecker interface {
	AssignmentExists(ctx context.Context, domainID, capabilityID string) (bool, error)
}

type EffectiveImportanceProjector struct {
	effectiveReadModel     *readmodels.EffectiveCapabilityImportanceReadModel
	importanceReadModel    *readmodels.StrategyImportanceReadModel
	capabilityReadModel    *readmodels.CapabilityReadModel
	domainAssignmentReader DomainAssignmentChecker
	ratingResolver         services.HierarchicalRatingResolver
	hierarchyService       services.CapabilityHierarchyService
	pillarsGateway         metamodel.StrategyPillarsGateway
}

func NewEffectiveImportanceProjector(
	effectiveReadModel *readmodels.EffectiveCapabilityImportanceReadModel,
	importanceReadModel *readmodels.StrategyImportanceReadModel,
	capabilityReadModel *readmodels.CapabilityReadModel,
	domainAssignmentReader DomainAssignmentChecker,
	ratingResolver services.HierarchicalRatingResolver,
	hierarchyService services.CapabilityHierarchyService,
	pillarsGateway metamodel.StrategyPillarsGateway,
) *EffectiveImportanceProjector {
	return &EffectiveImportanceProjector{
		effectiveReadModel:     effectiveReadModel,
		importanceReadModel:    importanceReadModel,
		capabilityReadModel:    capabilityReadModel,
		domainAssignmentReader: domainAssignmentReader,
		ratingResolver:         ratingResolver,
		hierarchyService:       hierarchyService,
		pillarsGateway:         pillarsGateway,
	}
}

func (p *EffectiveImportanceProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *EffectiveImportanceProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		"StrategyImportanceSet":          p.handleStrategyImportanceSet,
		"StrategyImportanceUpdated":      p.handleStrategyImportanceUpdated,
		"StrategyImportanceRemoved":      p.handleStrategyImportanceRemoved,
		"CapabilityParentChanged":        p.handleCapabilityParentChanged,
		"CapabilityDeleted":              p.handleCapabilityDeleted,
		"CapabilityAssignedToDomain":     p.handleCapabilityAssignedToDomain,
		"CapabilityUnassignedFromDomain": p.handleCapabilityUnassignedFromDomain,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

func (p *EffectiveImportanceProjector) handleStrategyImportanceSet(ctx context.Context, eventData []byte) error {
	var event events.StrategyImportanceSet
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal StrategyImportanceSet event: %v", err)
		return err
	}

	return p.recomputeCapabilityAndDescendants(ctx, event.CapabilityID, event.PillarID, event.BusinessDomainID)
}

func (p *EffectiveImportanceProjector) handleStrategyImportanceUpdated(ctx context.Context, eventData []byte) error {
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

	return p.recomputeCapabilityAndDescendants(ctx, importance.CapabilityID, importance.PillarID, importance.BusinessDomainID)
}

func (p *EffectiveImportanceProjector) handleStrategyImportanceRemoved(ctx context.Context, eventData []byte) error {
	var event events.StrategyImportanceRemoved
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal StrategyImportanceRemoved event: %v", err)
		return err
	}

	return p.recomputeCapabilityAndDescendants(ctx, event.CapabilityID, event.PillarID, event.BusinessDomainID)
}

func (p *EffectiveImportanceProjector) handleCapabilityParentChanged(ctx context.Context, eventData []byte) error {
	var event events.CapabilityParentChanged
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityParentChanged event: %v", err)
		return err
	}

	seen := make(map[string]bool)

	childEffectiveImportances, err := p.effectiveReadModel.GetByCapability(ctx, event.CapabilityID)
	if err != nil {
		log.Printf("Failed to get effective importances for capability %s: %v", event.CapabilityID, err)
		return err
	}

	for _, ei := range childEffectiveImportances {
		key := ei.PillarID + ":" + ei.BusinessDomainID
		if seen[key] {
			continue
		}
		seen[key] = true

		if err := p.recomputeCapabilityAndDescendants(ctx, event.CapabilityID, ei.PillarID, ei.BusinessDomainID); err != nil {
			log.Printf("Failed to recompute capability %s for pillar %s domain %s: %v",
				event.CapabilityID, ei.PillarID, ei.BusinessDomainID, err)
		}
	}

	if event.NewParentID != "" {
		parentEffectiveImportances, err := p.effectiveReadModel.GetByCapability(ctx, event.NewParentID)
		if err != nil {
			log.Printf("Failed to get effective importances for new parent %s: %v", event.NewParentID, err)
			return err
		}

		for _, ei := range parentEffectiveImportances {
			key := ei.PillarID + ":" + ei.BusinessDomainID
			if seen[key] {
				continue
			}
			seen[key] = true

			if err := p.recomputeCapabilityAndDescendants(ctx, event.CapabilityID, ei.PillarID, ei.BusinessDomainID); err != nil {
				log.Printf("Failed to recompute capability %s for pillar %s domain %s: %v",
					event.CapabilityID, ei.PillarID, ei.BusinessDomainID, err)
			}
		}
	}

	return nil
}

func (p *EffectiveImportanceProjector) handleCapabilityDeleted(ctx context.Context, eventData []byte) error {
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

func (p *EffectiveImportanceProjector) handleCapabilityAssignedToDomain(ctx context.Context, eventData []byte) error {
	var event events.CapabilityAssignedToDomain
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityAssignedToDomain event: %v", err)
		return err
	}

	pillars, err := p.pillarsGateway.GetStrategyPillars(ctx)
	if err != nil {
		log.Printf("Failed to get strategy pillars: %v", err)
		return err
	}

	for _, pillar := range pillars.Pillars {
		if !pillar.Active {
			continue
		}
		if err := p.recomputeCapabilityAndDescendants(ctx, event.CapabilityID, pillar.ID, event.BusinessDomainID); err != nil {
			log.Printf("Failed to recompute capability %s for pillar %s domain %s: %v",
				event.CapabilityID, pillar.ID, event.BusinessDomainID, err)
		}
	}

	return nil
}

func (p *EffectiveImportanceProjector) handleCapabilityUnassignedFromDomain(ctx context.Context, eventData []byte) error {
	var event events.CapabilityUnassignedFromDomain
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityUnassignedFromDomain event: %v", err)
		return err
	}

	capabilityID, err := valueobjects.NewCapabilityIDFromString(event.CapabilityID)
	if err != nil {
		return err
	}

	l1AncestorID, err := p.findL1Ancestor(ctx, capabilityID)
	if err != nil {
		log.Printf("Failed to find L1 ancestor for capability %s: %v", event.CapabilityID, err)
		return err
	}

	ancestorStillInDomain := false
	if l1AncestorID != "" && l1AncestorID != event.CapabilityID {
		exists, err := p.domainAssignmentReader.AssignmentExists(ctx, event.BusinessDomainID, l1AncestorID)
		if err != nil {
			log.Printf("Failed to check domain assignment for L1 ancestor %s: %v", l1AncestorID, err)
		} else {
			ancestorStillInDomain = exists
		}
	}

	if ancestorStillInDomain {
		log.Printf("L1 ancestor %s still assigned to domain %s, recomputing effective importance for %s",
			l1AncestorID, event.BusinessDomainID, event.CapabilityID)
		pillars, err := p.pillarsGateway.GetStrategyPillars(ctx)
		if err != nil {
			log.Printf("Failed to get strategy pillars: %v", err)
			return err
		}

		for _, pillar := range pillars.Pillars {
			if !pillar.Active {
				continue
			}
			if err := p.recomputeCapabilityAndDescendants(ctx, event.CapabilityID, pillar.ID, event.BusinessDomainID); err != nil {
				log.Printf("Failed to recompute capability %s for pillar %s domain %s: %v",
					event.CapabilityID, pillar.ID, event.BusinessDomainID, err)
			}
		}
		return nil
	}

	if err := p.effectiveReadModel.Delete(ctx, event.CapabilityID, "", event.BusinessDomainID); err != nil {
		log.Printf("Failed to delete effective importance for capability %s: %v", event.CapabilityID, err)
	}

	descendants, err := p.hierarchyService.GetDescendants(ctx, capabilityID)
	if err != nil {
		log.Printf("Failed to get descendants for capability %s: %v", event.CapabilityID, err)
		return err
	}

	for _, descendantID := range descendants {
		if err := p.effectiveReadModel.Delete(ctx, descendantID.Value(), "", event.BusinessDomainID); err != nil {
			log.Printf("Failed to delete effective importance for descendant %s: %v", descendantID.Value(), err)
		}
	}

	return nil
}

func (p *EffectiveImportanceProjector) findL1Ancestor(ctx context.Context, capabilityID valueobjects.CapabilityID) (string, error) {
	currentID := capabilityID

	for {
		capability, err := p.capabilityReadModel.GetByID(ctx, currentID.Value())
		if err != nil {
			return "", err
		}
		if capability == nil {
			return currentID.Value(), nil
		}

		if capability.Level == "L1" {
			return capability.ID, nil
		}

		if capability.ParentID == "" {
			return capability.ID, nil
		}

		currentID, err = valueobjects.NewCapabilityIDFromString(capability.ParentID)
		if err != nil {
			return "", err
		}
	}
}

func (p *EffectiveImportanceProjector) recomputeCapabilityAndDescendants(ctx context.Context, capabilityID, pillarID, businessDomainID string) error {
	if err := p.recomputeSingleCapability(ctx, capabilityID, pillarID, businessDomainID); err != nil {
		return err
	}

	capID, err := valueobjects.NewCapabilityIDFromString(capabilityID)
	if err != nil {
		return err
	}

	descendants, err := p.hierarchyService.GetDescendants(ctx, capID)
	if err != nil {
		log.Printf("Failed to get descendants for capability %s: %v", capabilityID, err)
		return err
	}

	for _, descendantID := range descendants {
		if err := p.recomputeSingleCapability(ctx, descendantID.Value(), pillarID, businessDomainID); err != nil {
			log.Printf("Failed to recompute descendant %s: %v", descendantID.Value(), err)
		}
	}

	return nil
}

func (p *EffectiveImportanceProjector) recomputeSingleCapability(ctx context.Context, capabilityID, pillarID, businessDomainID string) error {
	capID, err := valueobjects.NewCapabilityIDFromString(capabilityID)
	if err != nil {
		return err
	}

	pillarVo, err := valueobjects.NewPillarIDFromString(pillarID)
	if err != nil {
		return err
	}

	domainVo, err := valueobjects.NewBusinessDomainIDFromString(businessDomainID)
	if err != nil {
		return err
	}

	resolved, err := p.ratingResolver.ResolveEffectiveImportance(ctx, capID, pillarVo, domainVo)
	if err != nil {
		log.Printf("Failed to resolve effective importance for capability %s: %v", capabilityID, err)
		return err
	}

	if resolved == nil {
		return p.effectiveReadModel.Delete(ctx, capabilityID, pillarID, businessDomainID)
	}

	dto := readmodels.EffectiveImportanceDTO{
		CapabilityID:         capabilityID,
		PillarID:             pillarID,
		BusinessDomainID:     businessDomainID,
		EffectiveImportance:  resolved.EffectiveImportance.Importance().Value(),
		ImportanceLabel:      resolved.EffectiveImportance.Importance().Label(),
		SourceCapabilityID:   resolved.EffectiveImportance.SourceCapabilityID().Value(),
		SourceCapabilityName: resolved.SourceCapabilityName,
		IsInherited:          resolved.EffectiveImportance.IsInherited(),
		ComputedAt:           time.Now().UTC(),
	}

	return p.effectiveReadModel.Upsert(ctx, dto)
}
