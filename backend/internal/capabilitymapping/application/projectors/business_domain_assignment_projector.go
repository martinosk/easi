package projectors

import (
	"context"
	"encoding/json"
	"log"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/shared/domain"
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
		log.Printf("Failed to marshal event data: %v", err)
		return err
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
		log.Printf("Failed to unmarshal CapabilityAssignedToDomain event: %v", err)
		return err
	}

	domain, err := p.domainReadModel.GetByID(ctx, event.BusinessDomainID)
	if err != nil {
		log.Printf("Failed to get business domain %s: %v", event.BusinessDomainID, err)
		return err
	}
	if domain == nil {
		log.Printf("Business domain not found: %s", event.BusinessDomainID)
		return nil
	}

	capability, err := p.capabilityReadModel.GetByID(ctx, event.CapabilityID)
	if err != nil {
		log.Printf("Failed to get capability %s: %v", event.CapabilityID, err)
		return err
	}
	if capability == nil {
		log.Printf("Capability not found: %s", event.CapabilityID)
		return nil
	}

	dto := readmodels.AssignmentDTO{
		AssignmentID:       event.ID,
		BusinessDomainID:   event.BusinessDomainID,
		BusinessDomainName: domain.Name,
		CapabilityID:       event.CapabilityID,
		CapabilityCode:     capability.ID,
		CapabilityName:     capability.Name,
		CapabilityLevel:    capability.Level,
		AssignedAt:         event.AssignedAt,
	}
	return p.assignmentReadModel.Insert(ctx, dto)
}

func (p *BusinessDomainAssignmentProjector) handleCapabilityUnassignedFromDomain(ctx context.Context, eventData []byte) error {
	var event events.CapabilityUnassignedFromDomain
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityUnassignedFromDomain event: %v", err)
		return err
	}
	return p.assignmentReadModel.Delete(ctx, event.ID)
}
