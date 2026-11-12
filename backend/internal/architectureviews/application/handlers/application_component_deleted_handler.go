package handlers

import (
	"context"
	"fmt"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/shared/cqrs"
	domain "easi/backend/internal/shared/eventsourcing"
)

type ViewsContainingComponentQuery interface {
	GetViewsContainingComponent(ctx context.Context, componentID string) ([]string, error)
}

type ApplicationComponentDeletedHandler struct {
	commandBus cqrs.CommandBus
	readModel  ViewsContainingComponentQuery
}

func NewApplicationComponentDeletedHandler(
	commandBus cqrs.CommandBus,
	readModel ViewsContainingComponentQuery,
) *ApplicationComponentDeletedHandler {
	return &ApplicationComponentDeletedHandler{
		commandBus: commandBus,
		readModel:  readModel,
	}
}

func (h *ApplicationComponentDeletedHandler) Handle(ctx context.Context, event domain.DomainEvent) error {
	componentID := event.AggregateID()

	viewIDs, err := h.readModel.GetViewsContainingComponent(ctx, componentID)
	if err != nil {
		return fmt.Errorf("find views containing deleted application %s: %w", componentID, err)
	}

	for _, viewID := range viewIDs {
		cmd := &commands.RemoveComponentFromView{ViewID: viewID, ComponentID: componentID}
		if _, err := h.commandBus.Dispatch(ctx, cmd); err != nil {
			return fmt.Errorf("remove deleted application %s from view %s: %w", componentID, viewID, err)
		}
	}

	return nil
}
