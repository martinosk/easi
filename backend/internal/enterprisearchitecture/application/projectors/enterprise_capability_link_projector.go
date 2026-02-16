package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/enterprisearchitecture/domain/events"
	domain "easi/backend/internal/shared/eventsourcing"
)

type EnterpriseCapabilityLinkProjector struct {
	readModel *readmodels.EnterpriseCapabilityLinkReadModel
}

func NewEnterpriseCapabilityLinkProjector(readModel *readmodels.EnterpriseCapabilityLinkReadModel) *EnterpriseCapabilityLinkProjector {
	return &EnterpriseCapabilityLinkProjector{
		readModel: readModel,
	}
}

func (p *EnterpriseCapabilityLinkProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *EnterpriseCapabilityLinkProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		"EnterpriseCapabilityLinked":   p.handleLinked,
		"EnterpriseCapabilityUnlinked": p.handleUnlinked,
		cmPL.CapabilityParentChanged:   p.handleCapabilityParentChanged,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

func (p *EnterpriseCapabilityLinkProjector) handleLinked(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, event events.EnterpriseCapabilityLinked) error {
		dto := readmodels.EnterpriseCapabilityLinkDTO{
			ID:                     event.ID,
			EnterpriseCapabilityID: event.EnterpriseCapabilityID,
			DomainCapabilityID:     event.DomainCapabilityID,
			LinkedBy:               event.LinkedBy,
			LinkedAt:               event.LinkedAt,
		}
		if err := p.readModel.Insert(ctx, dto); err != nil {
			return fmt.Errorf("project EnterpriseCapabilityLinked for link %s: %w", event.ID, err)
		}
		if err := p.computeBlocking(ctx, event.DomainCapabilityID, event.EnterpriseCapabilityID); err != nil {
			return fmt.Errorf("project EnterpriseCapabilityLinked blocking computation for link %s: %w", event.ID, err)
		}
		return nil
	})
}

func (p *EnterpriseCapabilityLinkProjector) handleUnlinked(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, event events.EnterpriseCapabilityUnlinked) error {
		if err := p.readModel.DeleteBlockingByBlocker(ctx, event.DomainCapabilityID); err != nil {
			return fmt.Errorf("project EnterpriseCapabilityUnlinked delete blocking for capability %s: %w", event.DomainCapabilityID, err)
		}
		if err := p.readModel.Delete(ctx, event.ID); err != nil {
			return fmt.Errorf("project EnterpriseCapabilityUnlinked delete link %s: %w", event.ID, err)
		}
		return nil
	})
}

func (p *EnterpriseCapabilityLinkProjector) handleCapabilityParentChanged(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, event capabilityParentChangedEvent) error {
		if err := p.recomputeBlockingForSubtree(ctx, event.CapabilityID); err != nil {
			return fmt.Errorf("project CapabilityParentChanged blocking recomputation for capability %s: %w", event.CapabilityID, err)
		}
		return nil
	})
}

func (p *EnterpriseCapabilityLinkProjector) recomputeBlockingForSubtree(ctx context.Context, capabilityID string) error {
	subtreeIDs, err := p.readModel.QueryHierarchy(ctx, capabilityID, readmodels.HierarchySubtree)
	if err != nil {
		wrappedErr := fmt.Errorf("load subtree capability ids for capability %s: %w", capabilityID, err)
		log.Printf("failed to get subtree for capability %s: %v", capabilityID, wrappedErr)
		return wrappedErr
	}

	links, err := p.readModel.GetLinksForCapabilities(ctx, subtreeIDs)
	if err != nil {
		wrappedErr := fmt.Errorf("load enterprise links for subtree rooted at capability %s: %w", capabilityID, err)
		log.Printf("failed to get links for subtree of capability %s: %v", capabilityID, wrappedErr)
		return wrappedErr
	}

	if len(links) == 0 {
		return nil
	}

	linkedCapabilityIDs := make([]string, len(links))
	for i, link := range links {
		linkedCapabilityIDs[i] = link.DomainCapabilityID
	}
	if err := p.readModel.DeleteBlockingForCapabilities(ctx, linkedCapabilityIDs); err != nil {
		wrappedErr := fmt.Errorf("delete old blocking records for subtree rooted at capability %s: %w", capabilityID, err)
		log.Printf("failed to delete old blocking records: %v", wrappedErr)
		return wrappedErr
	}

	for _, link := range links {
		if err := p.computeBlocking(ctx, link.DomainCapabilityID, link.EnterpriseCapabilityID); err != nil {
			log.Printf("Failed to recompute blocking for link %s: %v", link.ID, err)
			return fmt.Errorf("recompute blocking for link %s (domain %s enterprise %s): %w", link.ID, link.DomainCapabilityID, link.EnterpriseCapabilityID, err)
		}
	}

	return nil
}

type blockingContext struct {
	domainCapabilityID   string
	enterpriseCapabilityID string
	capabilityName       string
	enterpriseName       string
}

func (p *EnterpriseCapabilityLinkProjector) computeBlocking(ctx context.Context, domainCapabilityID, enterpriseCapabilityID string) error {
	bc, err := p.buildBlockingContext(ctx, domainCapabilityID, enterpriseCapabilityID)
	if err != nil {
		return fmt.Errorf("build blocking context for domain capability %s and enterprise capability %s: %w", domainCapabilityID, enterpriseCapabilityID, err)
	}

	if err := p.insertBlockingForAncestors(ctx, bc); err != nil {
		return fmt.Errorf("insert ancestor blocking records for domain capability %s and enterprise capability %s: %w", domainCapabilityID, enterpriseCapabilityID, err)
	}
	if err := p.insertBlockingForDescendants(ctx, bc); err != nil {
		return fmt.Errorf("insert descendant blocking records for domain capability %s and enterprise capability %s: %w", domainCapabilityID, enterpriseCapabilityID, err)
	}

	return nil
}

func (p *EnterpriseCapabilityLinkProjector) buildBlockingContext(ctx context.Context, domainCapabilityID, enterpriseCapabilityID string) (blockingContext, error) {
	capabilityName, err := p.readModel.QueryName(ctx, domainCapabilityID, readmodels.NameDomainCapability)
	if err != nil {
		log.Printf("Failed to get capability name for %s: %v", domainCapabilityID, err)
		return blockingContext{}, fmt.Errorf("build blocking context names for domain %s enterprise %s: %w", domainCapabilityID, enterpriseCapabilityID, err)
	}

	enterpriseName, err := p.readModel.QueryName(ctx, enterpriseCapabilityID, readmodels.NameEnterpriseCapability)
	if err != nil {
		log.Printf("Failed to get enterprise capability name for %s: %v", enterpriseCapabilityID, err)
		return blockingContext{}, fmt.Errorf("build blocking context names for domain %s enterprise %s: %w", domainCapabilityID, enterpriseCapabilityID, err)
	}

	return blockingContext{
		domainCapabilityID:     domainCapabilityID,
		enterpriseCapabilityID: enterpriseCapabilityID,
		capabilityName:         capabilityName,
		enterpriseName:         enterpriseName,
	}, nil
}

func (p *EnterpriseCapabilityLinkProjector) insertBlockingForAncestors(ctx context.Context, bc blockingContext) error {
	ancestorIDs, err := p.readModel.QueryHierarchy(ctx, bc.domainCapabilityID, readmodels.HierarchyAncestors)
	if err != nil {
		log.Printf("Failed to get ancestors for capability %s: %v", bc.domainCapabilityID, err)
		return fmt.Errorf("load ancestors for capability %s while computing blocking: %w", bc.domainCapabilityID, err)
	}
	return p.insertBlockingRecords(ctx, bc, ancestorIDs, false)
}

func (p *EnterpriseCapabilityLinkProjector) insertBlockingForDescendants(ctx context.Context, bc blockingContext) error {
	descendantIDs, err := p.readModel.QueryHierarchy(ctx, bc.domainCapabilityID, readmodels.HierarchyDescendants)
	if err != nil {
		log.Printf("Failed to get descendants for capability %s: %v", bc.domainCapabilityID, err)
		return fmt.Errorf("load descendants for capability %s while computing blocking: %w", bc.domainCapabilityID, err)
	}
	return p.insertBlockingRecords(ctx, bc, descendantIDs, true)
}

func (p *EnterpriseCapabilityLinkProjector) insertBlockingRecords(ctx context.Context, bc blockingContext, relativeIDs []string, isAncestor bool) error {
	for _, relativeID := range relativeIDs {
		blocking := readmodels.BlockingDTO{
			DomainCapabilityID:      relativeID,
			BlockedByCapabilityID:   bc.domainCapabilityID,
			BlockedByEnterpriseID:   bc.enterpriseCapabilityID,
			BlockedByCapabilityName: bc.capabilityName,
			BlockedByEnterpriseName: bc.enterpriseName,
			IsAncestor:              isAncestor,
		}
		if err := p.readModel.InsertBlocking(ctx, blocking); err != nil {
			log.Printf("Failed to insert blocking for %s: %v", relativeID, err)
			return fmt.Errorf("insert blocking record for relative capability %s blocked by %s: %w", relativeID, bc.domainCapabilityID, err)
		}
	}

	return nil
}
