package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"
	domain "easi/backend/internal/shared/eventsourcing"
)

type BusinessDomainAssignmentProjector struct {
	assignmentReadModel *readmodels.DomainCapabilityAssignmentReadModel
	domainReadModel     *readmodels.BusinessDomainReadModel
	capabilityReadModel *readmodels.CapabilityReadModel
}

func NewBusinessDomainAssignmentProjector(
	assignmentReadModel *readmodels.DomainCapabilityAssignmentReadModel,
	domainReadModel *readmodels.BusinessDomainReadModel,
	capabilityReadModel *readmodels.CapabilityReadModel,
) *BusinessDomainAssignmentProjector {
	return &BusinessDomainAssignmentProjector{
		assignmentReadModel: assignmentReadModel,
		domainReadModel:     domainReadModel,
		capabilityReadModel: capabilityReadModel,
	}
}

func (p *BusinessDomainAssignmentProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *BusinessDomainAssignmentProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		"CapabilityAssignedToDomain":     p.handleCapabilityAssignedToDomain,
		"CapabilityUnassignedFromDomain": p.handleCapabilityUnassignedFromDomain,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

func (p *BusinessDomainAssignmentProjector) handleCapabilityAssignedToDomain(ctx context.Context, eventData []byte) error {
	var event events.CapabilityAssignedToDomain
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal CapabilityAssignedToDomain event data: %w", err)
		log.Printf("failed to unmarshal CapabilityAssignedToDomain event: %v", wrappedErr)
		return wrappedErr
	}

	domain, err := p.domainReadModel.GetByID(ctx, event.BusinessDomainID)
	if err != nil {
		log.Printf("Failed to get business domain %s: %v", event.BusinessDomainID, err)
		return fmt.Errorf("load business domain %s for assignment %s: %w", event.BusinessDomainID, event.ID, err)
	}
	if domain == nil {
		log.Printf("Business domain not found: %s", event.BusinessDomainID)
		return fmt.Errorf("business domain %s not found while handling CapabilityAssignedToDomain", event.BusinessDomainID)
	}

	capability, err := p.capabilityReadModel.GetByID(ctx, event.CapabilityID)
	if err != nil {
		log.Printf("Failed to get capability %s: %v", event.CapabilityID, err)
		return fmt.Errorf("load capability %s for assignment %s: %w", event.CapabilityID, event.ID, err)
	}
	if capability == nil {
		log.Printf("Capability not found: %s", event.CapabilityID)
		return fmt.Errorf("capability %s not found while handling CapabilityAssignedToDomain", event.CapabilityID)
	}

	dto := readmodels.AssignmentDTO{
		AssignmentID:          event.ID,
		BusinessDomainID:      event.BusinessDomainID,
		BusinessDomainName:    domain.Name,
		CapabilityID:          event.CapabilityID,
		CapabilityName:        capability.Name,
		CapabilityDescription: capability.Description,
		CapabilityLevel:       capability.Level,
		AssignedAt:            event.AssignedAt,
	}
	if err := p.assignmentReadModel.Insert(ctx, dto); err != nil {
		return fmt.Errorf("project CapabilityAssignedToDomain assignment insert for assignment %s: %w", event.ID, err)
	}
	return nil
}

func (p *BusinessDomainAssignmentProjector) handleCapabilityUnassignedFromDomain(ctx context.Context, eventData []byte) error {
	var event events.CapabilityUnassignedFromDomain
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal CapabilityUnassignedFromDomain event data: %w", err)
		log.Printf("failed to unmarshal CapabilityUnassignedFromDomain event: %v", wrappedErr)
		return wrappedErr
	}
	if err := p.assignmentReadModel.Delete(ctx, event.ID); err != nil {
		return fmt.Errorf("project CapabilityUnassignedFromDomain assignment delete for assignment %s: %w", event.ID, err)
	}
	return nil
}
