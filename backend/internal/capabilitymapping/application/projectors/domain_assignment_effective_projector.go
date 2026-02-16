package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/capabilitymapping/domain/services"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	mmPL "easi/backend/internal/metamodel/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type DomainAssignmentChecker interface {
	AssignmentExists(ctx context.Context, domainID, capabilityID string) (bool, error)
}

type DomainAncestryChecker struct {
	hierarchy        services.CapabilityHierarchyService
	domainAssignment DomainAssignmentChecker
}

func NewDomainAncestryChecker(hierarchy services.CapabilityHierarchyService, domainAssignment DomainAssignmentChecker) *DomainAncestryChecker {
	return &DomainAncestryChecker{hierarchy: hierarchy, domainAssignment: domainAssignment}
}

func (c *DomainAncestryChecker) IsAncestorInDomain(ctx context.Context, capabilityID, businessDomainID string) (bool, error) {
	capID, err := valueobjects.NewCapabilityIDFromString(capabilityID)
	if err != nil {
		return false, fmt.Errorf("parse capability ID %s for ancestor domain check: %w", capabilityID, err)
	}

	l1AncestorID, err := c.hierarchy.FindL1Ancestor(ctx, capID)
	if err != nil {
		return false, fmt.Errorf("find L1 ancestor for capability %s: %w", capabilityID, err)
	}

	if l1AncestorID.Value() == "" || l1AncestorID.Value() == capabilityID {
		return false, nil
	}

	exists, err := c.domainAssignment.AssignmentExists(ctx, businessDomainID, l1AncestorID.Value())
	if err != nil {
		log.Printf("Failed to check domain assignment for L1 ancestor %s: %v", l1AncestorID.Value(), err)
		return false, fmt.Errorf("check L1 ancestor domain assignment for capability %s domain %s: %w", capabilityID, businessDomainID, err)
	}

	return exists, nil
}

type DomainAssignmentEffectiveProjector struct {
	recomputer      *EffectiveImportanceRecomputer
	ancestryChecker *DomainAncestryChecker
	pillarsGateway  mmPL.StrategyPillarsGateway
}

func NewDomainAssignmentEffectiveProjector(
	recomputer *EffectiveImportanceRecomputer,
	ancestryChecker *DomainAncestryChecker,
	pillarsGateway mmPL.StrategyPillarsGateway,
) *DomainAssignmentEffectiveProjector {
	return &DomainAssignmentEffectiveProjector{
		recomputer:      recomputer,
		ancestryChecker: ancestryChecker,
		pillarsGateway:  pillarsGateway,
	}
}

func (p *DomainAssignmentEffectiveProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
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
		wrappedErr := fmt.Errorf("unmarshal CapabilityAssignedToDomain event data: %w", err)
		log.Printf("failed to unmarshal CapabilityAssignedToDomain event: %v", wrappedErr)
		return wrappedErr
	}

	return p.recomputeForAllActivePillars(ctx, event.CapabilityID, event.BusinessDomainID)
}

func (p *DomainAssignmentEffectiveProjector) handleCapabilityUnassignedFromDomain(ctx context.Context, eventData []byte) error {
	var event events.CapabilityUnassignedFromDomain
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal CapabilityUnassignedFromDomain event data: %w", err)
		log.Printf("failed to unmarshal CapabilityUnassignedFromDomain event: %v", wrappedErr)
		return wrappedErr
	}

	ancestorInDomain, err := p.ancestryChecker.IsAncestorInDomain(ctx, event.CapabilityID, event.BusinessDomainID)
	if err != nil {
		return fmt.Errorf("check ancestor domain assignment for capability %s domain %s: %w", event.CapabilityID, event.BusinessDomainID, err)
	}

	if ancestorInDomain {
		return p.recomputeForAllActivePillars(ctx, event.CapabilityID, event.BusinessDomainID)
	}

	return p.recomputer.DeleteCapabilityAndDescendants(ctx, event.CapabilityID, event.BusinessDomainID)
}

func (p *DomainAssignmentEffectiveProjector) recomputeForAllActivePillars(ctx context.Context, capabilityID, businessDomainID string) error {
	pillars, err := p.pillarsGateway.GetStrategyPillars(ctx)
	if err != nil {
		log.Printf("Failed to get strategy pillars: %v", err)
		return fmt.Errorf("get strategy pillars while recomputing capability %s domain %s: %w", capabilityID, businessDomainID, err)
	}

	for _, pillar := range pillars.Pillars {
		if !pillar.Active {
			continue
		}
		if err := p.recomputer.RecomputeCapabilityAndDescendants(ctx, ImportanceScope{
			CapabilityID:     capabilityID,
			PillarID:         pillar.ID,
			BusinessDomainID: businessDomainID,
		}); err != nil {
			log.Printf("Failed to recompute capability %s for pillar %s domain %s: %v",
				capabilityID, pillar.ID, businessDomainID, err)
			return fmt.Errorf("recompute effective importance for capability %s pillar %s domain %s: %w", capabilityID, pillar.ID, businessDomainID, err)
		}
	}

	return nil
}
