package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/application/readmodels"
	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	"easi/backend/internal/shared/cqrs"
	domain "easi/backend/internal/shared/eventsourcing"
)

type ActiveDirectionFinder interface {
	GetActiveByEnterpriseCapabilityID(ctx context.Context, enterpriseCapabilityID string) (*readmodels.DirectionDTO, error)
}

type CommandDispatcher interface {
	Dispatch(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error)
}

type EnterpriseCapabilityDeletedReactor struct {
	directions ActiveDirectionFinder
	commands   CommandDispatcher
}

func NewEnterpriseCapabilityDeletedReactor(directions ActiveDirectionFinder, commandDispatcher CommandDispatcher) *EnterpriseCapabilityDeletedReactor {
	return &EnterpriseCapabilityDeletedReactor{directions: directions, commands: commandDispatcher}
}

func (r *EnterpriseCapabilityDeletedReactor) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return r.ProjectEvent(ctx, event.EventType(), eventData)
}

func (r *EnterpriseCapabilityDeletedReactor) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	if eventType != pl.EnterpriseCapabilityDeleted {
		return nil
	}
	return handleProjection(ctx, eventData, r.rejectActiveDirection)
}

type enterpriseCapabilityDeletedEvent struct {
	ID string `json:"id"`
}

func (r *EnterpriseCapabilityDeletedReactor) rejectActiveDirection(ctx context.Context, event enterpriseCapabilityDeletedEvent) error {
	direction, err := r.directions.GetActiveByEnterpriseCapabilityID(ctx, event.ID)
	if err != nil {
		return fmt.Errorf("load active direction for deleted enterprise capability %s: %w", event.ID, err)
	}
	if direction == nil {
		return nil
	}
	if _, err := r.commands.Dispatch(ctx, &commands.RejectDirection{DirectionID: direction.ID}); err != nil {
		return fmt.Errorf("reject direction %s for deleted enterprise capability %s: %w", direction.ID, event.ID, err)
	}
	return nil
}
