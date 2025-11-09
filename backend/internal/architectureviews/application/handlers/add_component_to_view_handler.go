package handlers

import (
	"context"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/domain/valueobjects"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

// AddComponentToViewHandler handles AddComponentToView commands
type AddComponentToViewHandler struct {
	repository *repositories.ArchitectureViewRepository
}

// NewAddComponentToViewHandler creates a new handler
func NewAddComponentToViewHandler(repository *repositories.ArchitectureViewRepository) *AddComponentToViewHandler {
	return &AddComponentToViewHandler{
		repository: repository,
	}
}

// Handle processes the command
func (h *AddComponentToViewHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(commands.AddComponentToView)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	// Load the view aggregate
	view, err := h.repository.GetByID(ctx, command.ViewID)
	if err != nil {
		return err
	}

	// Create position value object
	position := valueobjects.NewComponentPosition(command.X, command.Y)

	// Add component to view
	if err := view.AddComponent(command.ComponentID, position); err != nil {
		return err
	}

	// Save updated aggregate
	return h.repository.Save(ctx, view)
}
