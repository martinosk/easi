package projectors

import (
	"context"
	"encoding/json"
	"log"

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

type DomainAssignmentEffectiveProjector struct {
	recomputer         *EffectiveImportanceRecomputer
	effectiveReadModel *readmodels.EffectiveCapabilityImportanceReadModel
	hierarchyService   services.CapabilityHierarchyService
	domainAssignment   DomainAssignmentChecker
	pillarsGateway     metamodel.StrategyPillarsGateway
}

func NewDomainAssignmentEffectiveProjector(
	recomputer *EffectiveImportanceRecomputer,
	effectiveReadModel *readmodels.EffectiveCapabilityImportanceReadModel,
	hierarchyService services.CapabilityHierarchyService,
	domainAssignment DomainAssignmentChecker,
	pillarsGateway metamodel.StrategyPillarsGateway,
) *DomainAssignmentEffectiveProjector {
	return &DomainAssignmentEffectiveProjector{
		recomputer:         recomputer,
		effectiveReadModel: effectiveReadModel,
		hierarchyService:   hierarchyService,
		domainAssignment:   domainAssignment,
		pillarsGateway:     pillarsGateway,
	}
}

func (p *DomainAssignmentEffectiveProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *DomainAssignmentEffectiveProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		"CapabilityAssignedToDomain":     p.handleCapabilityAssignedToDomain,
		"CapabilityUnassignedFromDomain": p.handleCapabilityUnassignedFromDomain,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

func (p *DomainAssignmentEffectiveProjector) handleCapabilityAssignedToDomain(ctx context.Context, eventData []byte) error {
	var event events.CapabilityAssignedToDomain
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityAssignedToDomain event: %v", err)
		return err
	}

	return p.recomputeForAllActivePillars(ctx, event.CapabilityID, event.BusinessDomainID)
}

func (p *DomainAssignmentEffectiveProjector) handleCapabilityUnassignedFromDomain(ctx context.Context, eventData []byte) error {
	var event events.CapabilityUnassignedFromDomain
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityUnassignedFromDomain event: %v", err)
		return err
	}

	if p.isAncestorStillInDomain(ctx, event.CapabilityID, event.BusinessDomainID) {
		return p.recomputeForAllActivePillars(ctx, event.CapabilityID, event.BusinessDomainID)
	}

	return p.deleteEffectiveImportanceForTree(ctx, event.CapabilityID, event.BusinessDomainID)
}

func (p *DomainAssignmentEffectiveProjector) isAncestorStillInDomain(ctx context.Context, capabilityID, businessDomainID string) bool {
	capID, err := valueobjects.NewCapabilityIDFromString(capabilityID)
	if err != nil {
		return false
	}

	l1AncestorID, err := p.hierarchyService.FindL1Ancestor(ctx, capID)
	if err != nil {
		return false
	}

	if l1AncestorID.Value() == "" {
		return false
	}

	if l1AncestorID.Value() == capabilityID {
		return false
	}

	exists, err := p.domainAssignment.AssignmentExists(ctx, businessDomainID, l1AncestorID.Value())
	if err != nil {
		log.Printf("Failed to check domain assignment for L1 ancestor %s: %v", l1AncestorID.Value(), err)
		return false
	}

	return exists
}

func (p *DomainAssignmentEffectiveProjector) recomputeForAllActivePillars(ctx context.Context, capabilityID, businessDomainID string) error {
	pillars, err := p.pillarsGateway.GetStrategyPillars(ctx)
	if err != nil {
		log.Printf("Failed to get strategy pillars: %v", err)
		return err
	}

	for _, pillar := range pillars.Pillars {
		if !pillar.Active {
			continue
		}
		if err := p.recomputer.RecomputeCapabilityAndDescendants(ctx, capabilityID, pillar.ID, businessDomainID); err != nil {
			log.Printf("Failed to recompute capability %s for pillar %s domain %s: %v",
				capabilityID, pillar.ID, businessDomainID, err)
		}
	}

	return nil
}

func (p *DomainAssignmentEffectiveProjector) deleteEffectiveImportanceForTree(ctx context.Context, capabilityID, businessDomainID string) error {
	if err := p.effectiveReadModel.Delete(ctx, capabilityID, "", businessDomainID); err != nil {
		log.Printf("Failed to delete effective importance for capability %s: %v", capabilityID, err)
	}

	return p.recomputer.ForEachDescendant(ctx, capabilityID, func(descendantID string) {
		if err := p.effectiveReadModel.Delete(ctx, descendantID, "", businessDomainID); err != nil {
			log.Printf("Failed to delete effective importance for descendant %s: %v", descendantID, err)
		}
	})
}
