package handlers

import (
	"context"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type RemoveComponentFromViewHandler struct {
	repository *repositories.ArchitectureViewRepository
}

func NewRemoveComponentFromViewHandler(repository *repositories.ArchitectureViewRepository) *RemoveComponentFromViewHandler {
	return &RemoveComponentFromViewHandler{
		repository: repository,
	}
}

func (h *RemoveComponentFromViewHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.RemoveComponentFromView)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	view, err := h.repository.GetByID(ctx, command.ViewID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := view.RemoveComponent(command.ComponentID); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, view); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
