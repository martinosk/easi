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

	relations, err := h.relationReadModel.GetByEntityID(ctx, command.ID)
	if err != nil {
		log.Printf("Error querying relationships for acquired entity %s: %v", command.ID, err)
		return cqrs.EmptyResult(), err
	}

	for _, relation := range relations {
		deleteCmd := &commands.DeleteAcquiredViaRelationship{ID: relation.ID}
		if _, err := h.commandBus.Dispatch(ctx, deleteCmd); err != nil {
			log.Printf("Error cascading delete for relationship %s: %v", relation.ID, err)
			continue
		}
		log.Printf("Cascaded delete for acquired via relationship %s", relation.ID)
	}

	return cqrs.EmptyResult(), nil
}
