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

func (h *RemoveComponentFromViewHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.RemoveComponentFromView)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	view, err := h.repository.GetByID(ctx, command.ViewID)
	if err != nil {
		return err
	}

	if err := view.RemoveComponent(command.ComponentID); err != nil {
		return err
	}

	return h.repository.Save(ctx, view)
}
