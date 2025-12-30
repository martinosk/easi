package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type CreateApplicationComponentHandler struct {
	repository *repositories.ApplicationComponentRepository
}

func NewCreateApplicationComponentHandler(repository *repositories.ApplicationComponentRepository) *CreateApplicationComponentHandler {
	return &CreateApplicationComponentHandler{
		repository: repository,
	}
}

func (h *CreateApplicationComponentHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.CreateApplicationComponent)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	name, err := valueobjects.NewComponentName(command.Name)
	if err != nil {
		return err
	}

	description, err := valueobjects.NewDescription(command.Description)
	if err != nil {
		return err
	}

	component, err := aggregates.NewApplicationComponent(name, description)
	if err != nil {
		return err
	}

	// Set the ID on the command so the caller can retrieve it
	command.ID = component.ID()

	return h.repository.Save(ctx, component)
}
