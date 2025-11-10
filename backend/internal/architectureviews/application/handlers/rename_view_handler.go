package handlers

import (
	"context"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/domain/valueobjects"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

// RenameViewHandler handles RenameView commands
type RenameViewHandler struct {
	repository *repositories.ArchitectureViewRepository
}

// NewRenameViewHandler creates a new handler
func NewRenameViewHandler(repository *repositories.ArchitectureViewRepository) *RenameViewHandler {
	return &RenameViewHandler{
		repository: repository,
	}
}

// Handle processes the command
func (h *RenameViewHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.RenameView)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	// Load the aggregate
	view, err := h.repository.GetByID(ctx, command.ViewID)
	if err != nil {
		return err
	}

	// Create value object with validation
	newName, err := valueobjects.NewViewName(command.NewName)
	if err != nil {
		return err
	}

	// Rename the view
	if err := view.Rename(newName); err != nil {
		return err
	}

	// Save to repository
	return h.repository.Save(ctx, view)
}
