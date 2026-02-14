package handlers

import (
	"context"

	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/valuestreams/application/commands"
	domain "easi/backend/internal/shared/eventsourcing"
)

type CapabilityDeletedHandler struct {
	commandBus *cqrs.InMemoryCommandBus
}

func NewCapabilityDeletedHandler(commandBus *cqrs.InMemoryCommandBus) *CapabilityDeletedHandler {
	return &CapabilityDeletedHandler{commandBus: commandBus}
}

func (h *CapabilityDeletedHandler) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData := event.EventData()
	capabilityID, ok := eventData["id"].(string)
	if !ok || capabilityID == "" {
		return nil
	}

	_, err := h.commandBus.Dispatch(ctx, &commands.RemoveDeletedCapability{
		CapabilityID: capabilityID,
	})
	return err
}
