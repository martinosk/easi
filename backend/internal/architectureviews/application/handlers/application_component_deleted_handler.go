package handlers

import (
	"context"
	"log"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/application/readmodels"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/eventsourcing"
)

type ApplicationComponentDeletedHandler struct {
	commandBus cqrs.CommandBus
	readModel  *readmodels.ArchitectureViewReadModel
}

func NewApplicationComponentDeletedHandler(
	commandBus cqrs.CommandBus,
	readModel *readmodels.ArchitectureViewReadModel,
) *ApplicationComponentDeletedHandler {
	return &ApplicationComponentDeletedHandler{
		commandBus: commandBus,
		readModel:  readModel,
	}
}

func (h *ApplicationComponentDeletedHandler) Handle(ctx context.Context, event domain.DomainEvent) error {
	componentID := event.AggregateID()

	log.Printf("Handling ApplicationComponentDeleted for component %s", componentID)

	viewIDs, err := h.readModel.GetViewsContainingComponent(ctx, componentID)
	if err != nil {
		log.Printf("Error querying views containing component %s: %v", componentID, err)
		return err
	}

	log.Printf("Found %d views containing component %s", len(viewIDs), componentID)

	for _, viewID := range viewIDs {
		cmd := &commands.RemoveComponentFromView{
			ViewID:      viewID,
			ComponentID: componentID,
		}

		if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
			log.Printf("Error removing component %s from view %s: %v", componentID, viewID, err)
			continue
		}

		log.Printf("Removed component %s from view %s", componentID, viewID)
	}

	return nil
}
