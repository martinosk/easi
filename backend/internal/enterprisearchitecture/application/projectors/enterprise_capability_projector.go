package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/enterprisearchitecture/domain/events"
	domain "easi/backend/internal/shared/eventsourcing"
)

func handleProjection[T any](ctx context.Context, eventData []byte, fn func(context.Context, T) error) error {
	var event T
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("unmarshal %T event data: %w", event, err)
	}
	return fn(ctx, event)
}

type EnterpriseCapabilityProjector struct {
	readModel *readmodels.EnterpriseCapabilityReadModel
}

func NewEnterpriseCapabilityProjector(readModel *readmodels.EnterpriseCapabilityReadModel) *EnterpriseCapabilityProjector {
	return &EnterpriseCapabilityProjector{
		readModel: readModel,
	}
}

func (p *EnterpriseCapabilityProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *EnterpriseCapabilityProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		"EnterpriseCapabilityCreated":           p.handleCreated,
		"EnterpriseCapabilityUpdated":           p.handleUpdated,
		"EnterpriseCapabilityDeleted":           p.handleDeleted,
		"EnterpriseCapabilityLinked":            p.handleLinked,
		"EnterpriseCapabilityUnlinked":          p.handleUnlinked,
		"EnterpriseCapabilityTargetMaturitySet": p.handleTargetMaturitySet,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

func (p *EnterpriseCapabilityProjector) handleCreated(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, p.projectCreated)
}

func (p *EnterpriseCapabilityProjector) projectCreated(ctx context.Context, event events.EnterpriseCapabilityCreated) error {
	return p.readModel.Insert(ctx, readmodels.EnterpriseCapabilityDTO{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
		Category:    event.Category,
		Active:      event.Active,
		CreatedAt:   event.CreatedAt,
	})
}

func (p *EnterpriseCapabilityProjector) handleUpdated(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, p.projectUpdated)
}

func (p *EnterpriseCapabilityProjector) projectUpdated(ctx context.Context, event events.EnterpriseCapabilityUpdated) error {
	return p.readModel.Update(ctx, readmodels.UpdateCapabilityParams{
		ID: event.ID, Name: event.Name, Description: event.Description, Category: event.Category,
	})
}

func (p *EnterpriseCapabilityProjector) handleDeleted(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, event events.EnterpriseCapabilityDeleted) error {
		return p.readModel.Delete(ctx, event.ID)
	})
}

func (p *EnterpriseCapabilityProjector) handleLinked(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, event events.EnterpriseCapabilityLinked) error {
		if err := p.readModel.IncrementLinkCount(ctx, event.EnterpriseCapabilityID); err != nil {
			return fmt.Errorf("increment link count for enterprise capability %s: %w", event.EnterpriseCapabilityID, err)
		}
		return p.readModel.RecalculateDomainCount(ctx, event.EnterpriseCapabilityID)
	})
}

func (p *EnterpriseCapabilityProjector) handleUnlinked(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, event events.EnterpriseCapabilityUnlinked) error {
		if err := p.readModel.DecrementLinkCount(ctx, event.EnterpriseCapabilityID); err != nil {
			return fmt.Errorf("decrement link count for enterprise capability %s: %w", event.EnterpriseCapabilityID, err)
		}
		return p.readModel.RecalculateDomainCount(ctx, event.EnterpriseCapabilityID)
	})
}

func (p *EnterpriseCapabilityProjector) handleTargetMaturitySet(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, event events.EnterpriseCapabilityTargetMaturitySet) error {
		return p.readModel.UpdateTargetMaturity(ctx, event.ID, event.TargetMaturity)
	})
}
