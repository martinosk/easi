package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

type UpdateApplicationComponentRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.ApplicationComponent, error)
	Save(ctx context.Context, component *aggregates.ApplicationComponent) error
}

type UpdateApplicationComponentHandler struct {
	repository UpdateApplicationComponentRepository
}

func NewUpdateApplicationComponentHandler(repository UpdateApplicationComponentRepository) *UpdateApplicationComponentHandler {
	return &UpdateApplicationComponentHandler{
		repository: repository,
	}
}

func (h *UpdateApplicationComponentHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.UpdateApplicationComponent)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	component, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	name, err := valueobjects.NewComponentName(command.Name)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	description, err := valueobjects.NewDescription(command.Description)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := component.Update(name, description); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, component); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
