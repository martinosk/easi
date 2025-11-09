package handlers

import (
	"context"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/domain/valueobjects"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

// UpdateComponentPositionHandler handles UpdateComponentPosition commands
type UpdateComponentPositionHandler struct {
	repository *repositories.ArchitectureViewRepository
}

// NewUpdateComponentPositionHandler creates a new handler
func NewUpdateComponentPositionHandler(repository *repositories.ArchitectureViewRepository) *UpdateComponentPositionHandler {
	return &UpdateComponentPositionHandler{
		repository: repository,
	}
}

// Handle processes the command
func (h *UpdateComponentPositionHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(commands.UpdateComponentPosition)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	// Load the view aggregate
	view, err := h.repository.GetByID(ctx, command.ViewID)
	if err != nil {
		return err
	}

	// Create new position value object
	newPosition := valueobjects.NewComponentPosition(command.X, command.Y)

	// Update component position
	if err := view.UpdateComponentPosition(command.ComponentID, newPosition); err != nil {
		return err
	}

	// Save updated aggregate
	return h.repository.Save(ctx, view)
}
