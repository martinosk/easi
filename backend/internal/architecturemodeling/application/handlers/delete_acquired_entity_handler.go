package handlers

import (
	"context"
	"log"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type DeleteAcquiredEntityHandler struct {
	repository         *repositories.AcquiredEntityRepository
	relationReadModel  *readmodels.AcquiredViaRelationshipReadModel
	commandBus         cqrs.CommandBus
}

func NewDeleteAcquiredEntityHandler(
	repository *repositories.AcquiredEntityRepository,
	relationReadModel *readmodels.AcquiredViaRelationshipReadModel,
	commandBus cqrs.CommandBus,
) *DeleteAcquiredEntityHandler {
	return &DeleteAcquiredEntityHandler{
		repository:        repository,
		relationReadModel: relationReadModel,
		commandBus:        commandBus,
	}
}

func (h *DeleteAcquiredEntityHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.DeleteAcquiredEntity)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	entity, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if entity.IsDeleted() {
		return cqrs.EmptyResult(), nil
	}

	if err := entity.Delete(); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, entity); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), h.cascadeClear(ctx, command.ID)
}

func (h *DeleteAcquiredEntityHandler) cascadeClear(ctx context.Context, entityID string) error {
	relations, err := h.relationReadModel.GetByEntityID(ctx, entityID)
	if err != nil {
		log.Printf("Error querying relationships for acquired entity %s: %v", entityID, err)
		return err
	}

	for _, relation := range relations {
		clearCmd := &commands.ClearOriginLink{ComponentID: relation.ComponentID, OriginType: "acquired-via"}
		if _, err := h.commandBus.Dispatch(ctx, clearCmd); err != nil {
			log.Printf("Error cascading clear for relationship on component %s: %v", relation.ComponentID, err)
			continue
		}
		log.Printf("Cascaded clear for acquired via relationship on component %s", relation.ComponentID)
	}
	return nil
}
