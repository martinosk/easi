package handlers

import (
	"context"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type DeleteViewHandler struct {
	repository *repositories.ArchitectureViewRepository
}

func NewDeleteViewHandler(repository *repositories.ArchitectureViewRepository) *DeleteViewHandler {
	return &DeleteViewHandler{
		repository: repository,
	}
}

func (h *DeleteViewHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.DeleteView)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	view, err := h.repository.GetByID(ctx, command.ViewID)
	if err != nil {
		return err
	}

	if err := view.Delete(); err != nil {
		return err
	}

	return h.repository.Save(ctx, view)
}
