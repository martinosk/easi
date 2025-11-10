package handlers

import (
	"context"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

// RemoveComponentFromViewHandler handles RemoveComponentFromView commands
type RemoveComponentFromViewHandler struct {
	repository *repositories.ArchitectureViewRepository
}

// NewRemoveComponentFromViewHandler creates a new handler
func NewRemoveComponentFromViewHandler(repository *repositories.ArchitectureViewRepository) *RemoveComponentFromViewHandler {
	return &RemoveComponentFromViewHandler{
		repository: repository,
	}
}

// Handle processes the command
func (h *RemoveComponentFromViewHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.RemoveComponentFromView)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	// Load the aggregate
	view, err := h.repository.GetByID(ctx, command.ViewID)
	if err != nil {
		return err
	}

	// Remove the component
	if err := view.RemoveComponent(command.ComponentID); err != nil {
		return err
	}

	// Save to repository
	return h.repository.Save(ctx, view)
}
