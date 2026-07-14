package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"easi/backend/internal/architecturedirection/application/commands"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

const clearedBySystemRealizationDeleted = "system:realization-deleted"

type RealizationRolePairFinder interface {
	FindPairByRealizationID(ctx context.Context, realizationID string) (string, string, bool, error)
}

type RealizationRoleDeletionReactor struct {
	pairs    RealizationRolePairFinder
	commands CommandDispatcher
}

func NewRealizationRoleDeletionReactor(pairs RealizationRolePairFinder, commandDispatcher CommandDispatcher) *RealizationRoleDeletionReactor {
	return &RealizationRoleDeletionReactor{pairs: pairs, commands: commandDispatcher}
}

func (r *RealizationRoleDeletionReactor) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return r.ProjectEvent(ctx, event.EventType(), eventData)
}

func (r *RealizationRoleDeletionReactor) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	if eventType != cmPL.SystemRealizationDeleted {
		return nil
	}
	var payload struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(eventData, &payload); err != nil {
		return fmt.Errorf("unmarshal SystemRealizationDeleted payload: %w", err)
	}
	return r.clearRoleForRealization(ctx, payload.ID)
}

func (r *RealizationRoleDeletionReactor) clearRoleForRealization(ctx context.Context, realizationID string) error {
	capabilityID, componentID, found, err := r.pairs.FindPairByRealizationID(ctx, realizationID)
	if err != nil {
		return fmt.Errorf("find realization role pair for deleted realization %s: %w", realizationID, err)
	}
	if !found {
		return nil
	}
	if _, err := r.commands.Dispatch(ctx, &commands.ClearRealizationRole{
		CapabilityID: capabilityID,
		ComponentID:  componentID,
		ClearedBy:    clearedBySystemRealizationDeleted,
	}); err != nil {
		return fmt.Errorf("clear realization role for deleted realization %s: %w", realizationID, err)
	}
	return nil
}
