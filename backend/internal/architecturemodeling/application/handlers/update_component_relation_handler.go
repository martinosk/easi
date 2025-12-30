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

func (h *UpdateComponentRelationHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.UpdateComponentRelation)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	relation, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	name, err := valueobjects.NewDescription(command.Name)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	description, err := valueobjects.NewDescription(command.Description)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := relation.Update(name, description); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, relation); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
