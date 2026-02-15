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
		log.Printf("Failed to marshal event data: %v", err)
		return err
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
	var event events.EnterpriseCapabilityLinked
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal EnterpriseCapabilityLinked event: %v", err)
		return err
	}

	dto := readmodels.EnterpriseCapabilityLinkDTO{
		ID:                     event.ID,
		EnterpriseCapabilityID: event.EnterpriseCapabilityID,
		DomainCapabilityID:     event.DomainCapabilityID,
		LinkedBy:               event.LinkedBy,
		LinkedAt:               event.LinkedAt,
	}
	if err := p.readModel.Insert(ctx, dto); err != nil {
		return err
	}

	return p.computeBlocking(ctx, event.DomainCapabilityID, event.EnterpriseCapabilityID)
}

func (p *EnterpriseCapabilityLinkProjector) handleUnlinked(ctx context.Context, eventData []byte) error {
	var event events.EnterpriseCapabilityUnlinked
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal EnterpriseCapabilityUnlinked event: %v", err)
		return err
	}

	if err := p.readModel.DeleteBlockingByBlocker(ctx, event.DomainCapabilityID); err != nil {
		log.Printf("Failed to delete blocking records for capability %s: %v", event.DomainCapabilityID, err)
		return err
	}

	return p.readModel.Delete(ctx, event.ID)
}

func (p *EnterpriseCapabilityLinkProjector) handleCapabilityParentChanged(ctx context.Context, eventData []byte) error {
	var event capabilityParentChangedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityParentChanged event: %v", err)
		return err
	}
	return p.recomputeBlockingForSubtree(ctx, event.CapabilityID)
}

func (p *EnterpriseCapabilityLinkProjector) recomputeBlockingForSubtree(ctx context.Context, capabilityID string) error {
	subtreeIDs, err := p.readModel.GetSubtreeCapabilityIDs(ctx, capabilityID)
	if err != nil {
		log.Printf("Failed to get subtree for capability %s: %v", capabilityID, err)
		return err
	}

	links, err := p.readModel.GetLinksForCapabilities(ctx, subtreeIDs)
	if err != nil {
		log.Printf("Failed to get links for subtree of capability %s: %v", capabilityID, err)
		return err
	}

	if len(links) == 0 {
		return nil
	}

	linkedCapabilityIDs := make([]string, len(links))
	for i, link := range links {
		linkedCapabilityIDs[i] = link.DomainCapabilityID
	}
	if err := p.readModel.DeleteBlockingForCapabilities(ctx, linkedCapabilityIDs); err != nil {
		log.Printf("Failed to delete old blocking records: %v", err)
		return err
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
		return err
	}

	if err := p.insertBlockingForAncestors(ctx, bc); err != nil {
		return err
	}
	if err := p.insertBlockingForDescendants(ctx, bc); err != nil {
		return err
	}

	return nil
}

func (p *EnterpriseCapabilityLinkProjector) buildBlockingContext(ctx context.Context, domainCapabilityID, enterpriseCapabilityID string) (blockingContext, error) {
	capabilityName, err := p.readModel.GetCapabilityName(ctx, domainCapabilityID)
	if err != nil {
		log.Printf("Failed to get capability name for %s: %v", domainCapabilityID, err)
		return blockingContext{}, fmt.Errorf("build blocking context names for domain %s enterprise %s: %w", domainCapabilityID, enterpriseCapabilityID, err)
	}

	enterpriseName, err := p.readModel.GetEnterpriseCapabilityName(ctx, enterpriseCapabilityID)
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
	ancestorIDs, err := p.readModel.GetAncestorIDs(ctx, bc.domainCapabilityID)
	if err != nil {
		log.Printf("Failed to get ancestors for capability %s: %v", bc.domainCapabilityID, err)
		return fmt.Errorf("load ancestors for capability %s while computing blocking: %w", bc.domainCapabilityID, err)
	}
	return p.insertBlockingRecords(ctx, bc, ancestorIDs, false)
}

func (p *EnterpriseCapabilityLinkProjector) insertBlockingForDescendants(ctx context.Context, bc blockingContext) error {
	descendantIDs, err := p.readModel.GetDescendantIDs(ctx, bc.domainCapabilityID)
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
