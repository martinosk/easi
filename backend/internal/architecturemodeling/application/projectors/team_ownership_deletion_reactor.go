package projectors

import (
	"context"
	"encoding/json"
	"fmt"

	"easi/backend/internal/architecturemodeling/application/commands"
	archPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	"easi/backend/internal/shared/cqrs"
	domain "easi/backend/internal/shared/eventsourcing"
)

type TeamOwnedComponentFinder interface {
	FindComponentIDsByTeamOwner(ctx context.Context, teamID string) ([]string, error)
}

type CommandDispatcher interface {
	Dispatch(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error)
}

type TeamOwnershipDeletionReactor struct {
	components TeamOwnedComponentFinder
	commands   CommandDispatcher
}

func NewTeamOwnershipDeletionReactor(components TeamOwnedComponentFinder, commandDispatcher CommandDispatcher) *TeamOwnershipDeletionReactor {
	return &TeamOwnershipDeletionReactor{components: components, commands: commandDispatcher}
}

func (r *TeamOwnershipDeletionReactor) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		return fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
	}
	return r.ProjectEvent(ctx, event.EventType(), eventData)
}

func (r *TeamOwnershipDeletionReactor) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	if eventType != archPL.InternalTeamDeleted {
		return nil
	}
	var payload struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(eventData, &payload); err != nil {
		return fmt.Errorf("unmarshal InternalTeamDeleted payload: %w", err)
	}
	return r.clearOwnershipForTeam(ctx, payload.ID)
}

func (r *TeamOwnershipDeletionReactor) clearOwnershipForTeam(ctx context.Context, teamID string) error {
	componentIDs, err := r.components.FindComponentIDsByTeamOwner(ctx, teamID)
	if err != nil {
		return fmt.Errorf("find components owned by deleted team %s: %w", teamID, err)
	}
	for _, componentID := range componentIDs {
		if _, err := r.commands.Dispatch(ctx, &commands.ClearApplicationComponentOwnership{ComponentID: componentID}); err != nil {
			return fmt.Errorf("clear ownership of component %s for deleted team %s: %w", componentID, teamID, err)
		}
	}
	return nil
}
