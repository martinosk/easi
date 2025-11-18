package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type DeleteComponentRelationHandler struct {
	repository *repositories.ComponentRelationRepository
}

func NewDeleteComponentRelationHandler(repository *repositories.ComponentRelationRepository) *DeleteComponentRelationHandler {
	return &DeleteComponentRelationHandler{
		repository: repository,
	}
}

func (h *DeleteComponentRelationHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.DeleteComponentRelation)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	relationID, err := valueobjects.NewRelationIDFromString(command.ID)
	if err != nil {
		return err
	}

	relation, err := h.repository.GetByID(ctx, relationID.Value())
	if err != nil {
		return err
	}

	if relation.IsDeleted() {
		return nil
	}

	if err := relation.Delete(); err != nil {
		return err
	}

	return h.repository.Save(ctx, relation)
}
