package handlers

import (
	"context"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

// DeleteViewHandler handles DeleteView commands
type DeleteViewHandler struct {
	repository *repositories.ArchitectureViewRepository
}

// NewDeleteViewHandler creates a new handler
func NewDeleteViewHandler(repository *repositories.ArchitectureViewRepository) *DeleteViewHandler {
	return &DeleteViewHandler{
		repository: repository,
	}
}

// Handle processes the command
func (h *DeleteViewHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.DeleteView)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	// Load the aggregate
	view, err := h.repository.GetByID(ctx, command.ViewID)
	if err != nil {
		return err
	}

	// Delete the view
	if err := view.Delete(); err != nil {
		return err
	}

	// Save to repository
	return h.repository.Save(ctx, view)
}
