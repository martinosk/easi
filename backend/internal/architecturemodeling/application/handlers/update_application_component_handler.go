package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type UpdateApplicationComponentHandler struct {
	repository *repositories.ApplicationComponentRepository
}

func NewUpdateApplicationComponentHandler(repository *repositories.ApplicationComponentRepository) *UpdateApplicationComponentHandler {
	return &UpdateApplicationComponentHandler{
		repository: repository,
	}
}

func (h *UpdateApplicationComponentHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.UpdateApplicationComponent)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	component, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	name, err := valueobjects.NewComponentName(command.Name)
	if err != nil {
		return err
	}

	description, err := valueobjects.NewDescription(command.Description)
	if err != nil {
		return err
	}

	if err := component.Update(name, description); err != nil {
		return err
	}

	return h.repository.Save(ctx, component)
}
