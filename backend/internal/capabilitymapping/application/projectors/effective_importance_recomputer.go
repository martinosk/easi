package projectors

import (
	"context"
	"fmt"
	"log"
	"time"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/capabilitymapping/domain/services"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	sharedEvents "easi/backend/internal/shared/events"
	domain "easi/backend/internal/shared/eventsourcing"
)

type ImportanceScope struct {
	CapabilityID     string
	PillarID         string
	BusinessDomainID string
}

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

func (r *EffectiveImportanceRecomputer) RecomputeCapabilityAndDescendants(ctx context.Context, scope ImportanceScope) error {
	if err := r.recomputeSingleCapability(ctx, scope); err != nil {
		return fmt.Errorf("recompute effective importance for capability %s pillar %s domain %s: %w", scope.CapabilityID, scope.PillarID, scope.BusinessDomainID, err)
	}

	return r.ForEachDescendant(ctx, scope.CapabilityID, func(descendantID string) error {
		descendantScope := ImportanceScope{
			CapabilityID:     descendantID,
			PillarID:         scope.PillarID,
			BusinessDomainID: scope.BusinessDomainID,
		}
		if err := r.recomputeSingleCapability(ctx, descendantScope); err != nil {
			log.Printf("Failed to recompute descendant %s: %v", descendantID, err)
			return fmt.Errorf("recompute descendant %s (root %s, pillar %s, domain %s): %w",
				descendantID, scope.CapabilityID, scope.PillarID, scope.BusinessDomainID, err)
		}
		return nil
	})
}

func (r *EffectiveImportanceRecomputer) ForEachDescendant(ctx context.Context, capabilityID string, fn func(descendantID string) error) error {
	capID, err := valueobjects.NewCapabilityIDFromString(capabilityID)
	if err != nil {
		return fmt.Errorf("parse capability id %s for descendant traversal: %w", capabilityID, err)
	}

	descendants, err := r.hierarchyService.GetDescendants(ctx, capID)
	if err != nil {
		log.Printf("Failed to get descendants for capability %s: %v", capabilityID, err)
		return fmt.Errorf("load descendants for capability %s: %w", capabilityID, err)
	}

	for _, descendantID := range descendants {
		if err := fn(descendantID.Value()); err != nil {
			return fmt.Errorf("process descendant %s of capability %s: %w", descendantID.Value(), capabilityID, err)
		}
	}

	return nil
}

func (r *EffectiveImportanceRecomputer) recomputeSingleCapability(ctx context.Context, scope ImportanceScope) error {
	resolved, err := r.resolveImportance(ctx, scope)
	if err != nil {
		return fmt.Errorf("resolve effective importance for capability %s pillar %s domain %s: %w", scope.CapabilityID, scope.PillarID, scope.BusinessDomainID, err)
	}

	if resolved == nil {
		if err := r.effectiveReadModel.Delete(ctx, scope.CapabilityID, scope.PillarID, scope.BusinessDomainID); err != nil {
			return fmt.Errorf("delete effective importance for capability %s pillar %s domain %s: %w", scope.CapabilityID, scope.PillarID, scope.BusinessDomainID, err)
		}
		return nil
	}

	dto := r.buildDTO(scope, resolved)
	if err := r.effectiveReadModel.Upsert(ctx, dto); err != nil {
		return fmt.Errorf("upsert effective importance for capability %s pillar %s domain %s: %w", scope.CapabilityID, scope.PillarID, scope.BusinessDomainID, err)
	}

	return r.publishRecalculatedEvent(ctx, dto)
}

func (r *EffectiveImportanceRecomputer) resolveImportance(ctx context.Context, scope ImportanceScope) (*services.ResolvedRating, error) {
	capID, err := valueobjects.NewCapabilityIDFromString(scope.CapabilityID)
	if err != nil {
		return nil, fmt.Errorf("parse capability id %s for effective importance resolution: %w", scope.CapabilityID, err)
	}

	pillarVo, err := valueobjects.NewPillarIDFromString(scope.PillarID)
	if err != nil {
		return nil, fmt.Errorf("parse pillar id %s for effective importance resolution: %w", scope.PillarID, err)
	}

	domainVo, err := valueobjects.NewBusinessDomainIDFromString(scope.BusinessDomainID)
	if err != nil {
		return nil, fmt.Errorf("parse business domain id %s for effective importance resolution: %w", scope.BusinessDomainID, err)
	}

	resolved, err := r.ratingResolver.ResolveEffectiveImportance(ctx, capID, pillarVo, domainVo)
	if err != nil {
		log.Printf("Failed to resolve effective importance for capability %s: %v", scope.CapabilityID, err)
		return nil, fmt.Errorf("resolve effective importance via resolver for capability %s pillar %s domain %s: %w", scope.CapabilityID, scope.PillarID, scope.BusinessDomainID, err)
	}
	return resolved, nil
}

func (r *EffectiveImportanceRecomputer) buildDTO(scope ImportanceScope, resolved *services.ResolvedRating) readmodels.EffectiveImportanceDTO {
	return readmodels.EffectiveImportanceDTO{
		CapabilityID:         scope.CapabilityID,
		PillarID:             scope.PillarID,
		BusinessDomainID:     scope.BusinessDomainID,
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
		return fmt.Errorf("delete effective importance for capability %s domain %s: %w", capabilityID, businessDomainID, err)
	}

	return r.ForEachDescendant(ctx, capabilityID, func(descendantID string) error {
		if err := r.effectiveReadModel.Delete(ctx, descendantID, "", businessDomainID); err != nil {
			log.Printf("Failed to delete effective importance for descendant %s: %v", descendantID, err)
			return fmt.Errorf("delete effective importance for descendant %s domain %s: %w", descendantID, businessDomainID, err)
		}
		return nil
	})
}

func (r *EffectiveImportanceRecomputer) publishRecalculatedEvent(ctx context.Context, dto readmodels.EffectiveImportanceDTO) error {
	if r.eventBus == nil {
		return nil
	}
	recalcEvent := events.NewEffectiveImportanceRecalculated(
		dto.CapabilityID, dto.BusinessDomainID, dto.PillarID, dto.EffectiveImportance,
	)
	if err := r.eventBus.Publish(ctx, []domain.DomainEvent{recalcEvent}); err != nil {
		log.Printf("Failed to publish EffectiveImportanceRecalculated event for capability %s: %v", dto.CapabilityID, err)
		return fmt.Errorf("publish EffectiveImportanceRecalculated for capability %s pillar %s domain %s: %w",
			dto.CapabilityID, dto.PillarID, dto.BusinessDomainID, err)
	}

	return nil
}
