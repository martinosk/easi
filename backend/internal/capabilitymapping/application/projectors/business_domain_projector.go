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
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
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
	return handleProjection(ctx, eventData, p.projectBusinessDomainCreated)
}

func (p *BusinessDomainProjector) projectBusinessDomainCreated(ctx context.Context, event events.BusinessDomainCreated) error {
	return p.readModel.Insert(ctx, readmodels.BusinessDomainDTO{
		ID:                event.ID,
		Name:              event.Name,
		Description:       event.Description,
		DomainArchitectID: event.DomainArchitectID,
		CreatedAt:         event.CreatedAt,
	})
}

func (p *BusinessDomainProjector) handleBusinessDomainUpdated(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, p.projectBusinessDomainUpdated)
}

func (p *BusinessDomainProjector) projectBusinessDomainUpdated(ctx context.Context, event events.BusinessDomainUpdated) error {
	return p.readModel.Update(ctx, event.ID, readmodels.BusinessDomainUpdate{
		Name:              event.Name,
		Description:       event.Description,
		DomainArchitectID: event.DomainArchitectID,
	})
}

func (p *BusinessDomainProjector) handleBusinessDomainDeleted(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, event events.BusinessDomainDeleted) error {
		return p.readModel.Delete(ctx, event.ID)
	})
}

func (p *BusinessDomainProjector) handleCapabilityAssignedToDomain(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, event events.CapabilityAssignedToDomain) error {
		return p.readModel.IncrementCapabilityCount(ctx, event.BusinessDomainID)
	})
}

func (p *BusinessDomainProjector) handleCapabilityUnassignedFromDomain(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, event events.CapabilityUnassignedFromDomain) error {
		return p.readModel.DecrementCapabilityCount(ctx, event.BusinessDomainID)
	})
}
