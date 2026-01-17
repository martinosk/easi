package projectors

import (
	"context"
	"encoding/json"
	"log"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"
	domain "easi/backend/internal/shared/eventsourcing"
)

type BusinessDomainProjector struct {
	readModel *readmodels.BusinessDomainReadModel
}

func NewBusinessDomainProjector(readModel *readmodels.BusinessDomainReadModel) *BusinessDomainProjector {
	return &BusinessDomainProjector{
		readModel: readModel,
	}
}

func (p *BusinessDomainProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *BusinessDomainProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		"BusinessDomainCreated":          p.handleBusinessDomainCreated,
		"BusinessDomainUpdated":          p.handleBusinessDomainUpdated,
		"BusinessDomainDeleted":          p.handleBusinessDomainDeleted,
		"CapabilityAssignedToDomain":     p.handleCapabilityAssignedToDomain,
		"CapabilityUnassignedFromDomain": p.handleCapabilityUnassignedFromDomain,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

func (p *BusinessDomainProjector) handleBusinessDomainCreated(ctx context.Context, eventData []byte) error {
	var event events.BusinessDomainCreated
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal BusinessDomainCreated event: %v", err)
		return err
	}

	dto := readmodels.BusinessDomainDTO{
		ID:                event.ID,
		Name:              event.Name,
		Description:       event.Description,
		DomainArchitectID: event.DomainArchitectID,
		CreatedAt:         event.CreatedAt,
	}
	return p.readModel.Insert(ctx, dto)
}

func (p *BusinessDomainProjector) handleBusinessDomainUpdated(ctx context.Context, eventData []byte) error {
	var event events.BusinessDomainUpdated
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal BusinessDomainUpdated event: %v", err)
		return err
	}
	return p.readModel.Update(ctx, event.ID, event.Name, event.Description, event.DomainArchitectID)
}

func (p *BusinessDomainProjector) handleBusinessDomainDeleted(ctx context.Context, eventData []byte) error {
	var event events.BusinessDomainDeleted
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal BusinessDomainDeleted event: %v", err)
		return err
	}
	return p.readModel.Delete(ctx, event.ID)
}

func (p *BusinessDomainProjector) handleCapabilityAssignedToDomain(ctx context.Context, eventData []byte) error {
	var event events.CapabilityAssignedToDomain
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityAssignedToDomain event: %v", err)
		return err
	}
	return p.readModel.IncrementCapabilityCount(ctx, event.BusinessDomainID)
}

func (p *BusinessDomainProjector) handleCapabilityUnassignedFromDomain(ctx context.Context, eventData []byte) error {
	var event events.CapabilityUnassignedFromDomain
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityUnassignedFromDomain event: %v", err)
		return err
	}
	return p.readModel.DecrementCapabilityCount(ctx, event.BusinessDomainID)
}
