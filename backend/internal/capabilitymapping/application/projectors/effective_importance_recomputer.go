package projectors

import (
	"context"
	"log"
	"time"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/services"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
)

type EffectiveImportanceRecomputer struct {
	effectiveReadModel *readmodels.EffectiveCapabilityImportanceReadModel
	ratingResolver     services.HierarchicalRatingResolver
	hierarchyService   services.CapabilityHierarchyService
}

func NewEffectiveImportanceRecomputer(
	effectiveReadModel *readmodels.EffectiveCapabilityImportanceReadModel,
	ratingResolver services.HierarchicalRatingResolver,
	hierarchyService services.CapabilityHierarchyService,
) *EffectiveImportanceRecomputer {
	return &EffectiveImportanceRecomputer{
		effectiveReadModel: effectiveReadModel,
		ratingResolver:     ratingResolver,
		hierarchyService:   hierarchyService,
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

	resolved, err := r.ratingResolver.ResolveEffectiveImportance(ctx, capID, pillarVo, domainVo)
	if err != nil {
		log.Printf("Failed to resolve effective importance for capability %s: %v", capabilityID, err)
		return err
	}

	if resolved == nil {
		return r.effectiveReadModel.Delete(ctx, capabilityID, pillarID, businessDomainID)
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
		Rationale:            resolved.Rationale,
		ComputedAt:           time.Now().UTC(),
	}

	return r.effectiveReadModel.Upsert(ctx, dto)
}
