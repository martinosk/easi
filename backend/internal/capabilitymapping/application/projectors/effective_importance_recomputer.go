package projectors

import (
	"context"
	"log"
	"time"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/capabilitymapping/domain/services"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedEvents "easi/backend/internal/shared/events"
)

type EffectiveImportanceRecomputer struct {
	effectiveReadModel *readmodels.EffectiveCapabilityImportanceReadModel
	ratingResolver     services.HierarchicalRatingResolver
	hierarchyService   services.CapabilityHierarchyService
	eventBus           sharedEvents.EventBus
}

func NewEffectiveImportanceRecomputer(
	effectiveReadModel *readmodels.EffectiveCapabilityImportanceReadModel,
	ratingResolver services.HierarchicalRatingResolver,
	hierarchyService services.CapabilityHierarchyService,
	eventBus sharedEvents.EventBus,
) *EffectiveImportanceRecomputer {
	return &EffectiveImportanceRecomputer{
		effectiveReadModel: effectiveReadModel,
		ratingResolver:     ratingResolver,
		hierarchyService:   hierarchyService,
		eventBus:           eventBus,
	}
}

func (r *EffectiveImportanceRecomputer) RecomputeCapabilityAndDescendants(ctx context.Context, capabilityID, pillarID, businessDomainID string) error {
	if err := r.recomputeSingleCapability(ctx, capabilityID, pillarID, businessDomainID); err != nil {
		return err
	}

	return r.ForEachDescendant(ctx, capabilityID, func(descendantID string) {
		if err := r.recomputeSingleCapability(ctx, descendantID, pillarID, businessDomainID); err != nil {
			log.Printf("Failed to recompute descendant %s: %v", descendantID, err)
		}
	})
}

func (r *EffectiveImportanceRecomputer) ForEachDescendant(ctx context.Context, capabilityID string, fn func(descendantID string)) error {
	capID, err := valueobjects.NewCapabilityIDFromString(capabilityID)
	if err != nil {
		return err
	}

	descendants, err := r.hierarchyService.GetDescendants(ctx, capID)
	if err != nil {
		log.Printf("Failed to get descendants for capability %s: %v", capabilityID, err)
		return err
	}

	for _, descendantID := range descendants {
		fn(descendantID.Value())
	}

	return nil
}

func (r *EffectiveImportanceRecomputer) recomputeSingleCapability(ctx context.Context, capabilityID, pillarID, businessDomainID string) error {
	resolved, err := r.resolveImportance(ctx, capabilityID, pillarID, businessDomainID)
	if err != nil {
		return err
	}

	if resolved == nil {
		return r.effectiveReadModel.Delete(ctx, capabilityID, pillarID, businessDomainID)
	}

	dto := r.buildDTO(capabilityID, pillarID, businessDomainID, resolved)
	if err := r.effectiveReadModel.Upsert(ctx, dto); err != nil {
		return err
	}

	r.publishRecalculatedEvent(ctx, dto)
	return nil
}

func (r *EffectiveImportanceRecomputer) resolveImportance(ctx context.Context, capabilityID, pillarID, businessDomainID string) (*services.ResolvedRating, error) {
	capID, err := valueobjects.NewCapabilityIDFromString(capabilityID)
	if err != nil {
		return nil, err
	}

	pillarVo, err := valueobjects.NewPillarIDFromString(pillarID)
	if err != nil {
		return nil, err
	}

	domainVo, err := valueobjects.NewBusinessDomainIDFromString(businessDomainID)
	if err != nil {
		return nil, err
	}

	resolved, err := r.ratingResolver.ResolveEffectiveImportance(ctx, capID, pillarVo, domainVo)
	if err != nil {
		log.Printf("Failed to resolve effective importance for capability %s: %v", capabilityID, err)
		return nil, err
	}
	return resolved, nil
}

func (r *EffectiveImportanceRecomputer) buildDTO(capabilityID, pillarID, businessDomainID string, resolved *services.ResolvedRating) readmodels.EffectiveImportanceDTO {
	return readmodels.EffectiveImportanceDTO{
		CapabilityID:         capabilityID,
		PillarID:             pillarID,
		BusinessDomainID:     businessDomainID,
		EffectiveImportance:  resolved.EffectiveImportance.Importance().Value(),
		ImportanceLabel:      resolved.EffectiveImportance.Importance().Label(),
		SourceCapabilityID:   resolved.EffectiveImportance.SourceCapabilityID().Value(),
		SourceCapabilityName: resolved.SourceCapabilityName,
		IsInherited:          resolved.EffectiveImportance.IsInherited(),
		Rationale:            resolved.Rationale,
		ComputedAt:           time.Now().UTC(),
	}
}

func (r *EffectiveImportanceRecomputer) DeleteCapabilityAndDescendants(ctx context.Context, capabilityID, businessDomainID string) error {
	if err := r.effectiveReadModel.Delete(ctx, capabilityID, "", businessDomainID); err != nil {
		log.Printf("Failed to delete effective importance for capability %s: %v", capabilityID, err)
	}

	return r.ForEachDescendant(ctx, capabilityID, func(descendantID string) {
		if err := r.effectiveReadModel.Delete(ctx, descendantID, "", businessDomainID); err != nil {
			log.Printf("Failed to delete effective importance for descendant %s: %v", descendantID, err)
		}
	})
}

func (r *EffectiveImportanceRecomputer) publishRecalculatedEvent(ctx context.Context, dto readmodels.EffectiveImportanceDTO) {
	if r.eventBus == nil {
		return
	}
	recalcEvent := events.NewEffectiveImportanceRecalculated(
		dto.CapabilityID, dto.BusinessDomainID, dto.PillarID, dto.EffectiveImportance,
	)
	if err := r.eventBus.Publish(ctx, []domain.DomainEvent{recalcEvent}); err != nil {
		log.Printf("Failed to publish EffectiveImportanceRecalculated event for capability %s: %v", dto.CapabilityID, err)
	}
}
