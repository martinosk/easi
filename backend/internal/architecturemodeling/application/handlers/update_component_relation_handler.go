package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type UpdateComponentRelationHandler struct {
	repository *repositories.ComponentRelationRepository
}

func NewUpdateComponentRelationHandler(repository *repositories.ComponentRelationRepository) *UpdateComponentRelationHandler {
	return &UpdateComponentRelationHandler{
		repository: repository,
	}
}

func (h *UpdateComponentRelationHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.UpdateComponentRelation)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	relation, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	name, err := valueobjects.NewDescription(command.Name)
	if err != nil {
		return err
	}
	description, err := valueobjects.NewDescription(command.Description)
	if err != nil {
		return err
	}

	if err := relation.Update(name, description); err != nil {
		return err
	}

	return h.repository.Save(ctx, relation)
}
