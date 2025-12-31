package projectors

import (
	"context"
	"encoding/json"
	"log"

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
		"CapabilityParentChanged":      p.handleCapabilityParentChanged,
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

	subtreeIDs, err := p.readModel.GetSubtreeCapabilityIDs(ctx, event.CapabilityID)
	if err != nil {
		log.Printf("Failed to get subtree for capability %s: %v", event.CapabilityID, err)
		return err
	}

	links, err := p.readModel.GetLinksForCapabilities(ctx, subtreeIDs)
	if err != nil {
		log.Printf("Failed to get links for subtree of capability %s: %v", event.CapabilityID, err)
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
		}
	}

	return nil
}

func (p *EnterpriseCapabilityLinkProjector) computeBlocking(ctx context.Context, domainCapabilityID, enterpriseCapabilityID string) error {
	capabilityName := p.getCapabilityNameOrDefault(ctx, domainCapabilityID)
	enterpriseName := p.getEnterpriseNameOrDefault(ctx, enterpriseCapabilityID)

	p.insertBlockingForRelatives(ctx, domainCapabilityID, enterpriseCapabilityID, capabilityName, enterpriseName, false)
	p.insertBlockingForRelatives(ctx, domainCapabilityID, enterpriseCapabilityID, capabilityName, enterpriseName, true)

	return nil
}

func (p *EnterpriseCapabilityLinkProjector) getCapabilityNameOrDefault(ctx context.Context, capabilityID string) string {
	name, err := p.readModel.GetCapabilityName(ctx, capabilityID)
	if err != nil {
		log.Printf("Failed to get capability name for %s: %v", capabilityID, err)
		return capabilityID
	}
	return name
}

func (p *EnterpriseCapabilityLinkProjector) getEnterpriseNameOrDefault(ctx context.Context, enterpriseCapabilityID string) string {
	name, err := p.readModel.GetEnterpriseCapabilityName(ctx, enterpriseCapabilityID)
	if err != nil {
		log.Printf("Failed to get enterprise capability name for %s: %v", enterpriseCapabilityID, err)
		return enterpriseCapabilityID
	}
	return name
}

func (p *EnterpriseCapabilityLinkProjector) insertBlockingForRelatives(ctx context.Context, domainCapabilityID, enterpriseCapabilityID, capabilityName, enterpriseName string, isDescendants bool) {
	var relativeIDs []string
	var err error
	var relationType string

	if isDescendants {
		relativeIDs, err = p.readModel.GetDescendantIDs(ctx, domainCapabilityID)
		relationType = "descendants"
	} else {
		relativeIDs, err = p.readModel.GetAncestorIDs(ctx, domainCapabilityID)
		relationType = "ancestors"
	}

	if err != nil {
		log.Printf("Failed to get %s for capability %s: %v", relationType, domainCapabilityID, err)
		return
	}

	for _, relativeID := range relativeIDs {
		blocking := readmodels.BlockingDTO{
			DomainCapabilityID:      relativeID,
			BlockedByCapabilityID:   domainCapabilityID,
			BlockedByEnterpriseID:   enterpriseCapabilityID,
			BlockedByCapabilityName: capabilityName,
			BlockedByEnterpriseName: enterpriseName,
			IsAncestor:              isDescendants,
		}
		if err := p.readModel.InsertBlocking(ctx, blocking); err != nil {
			log.Printf("Failed to insert blocking for %s %s: %v", relationType, relativeID, err)
		}
	}
}
