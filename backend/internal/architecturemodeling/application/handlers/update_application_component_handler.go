package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

// UpdateApplicationComponentHandler handles UpdateApplicationComponent commands
type UpdateApplicationComponentHandler struct {
	repository *repositories.ApplicationComponentRepository
}

// NewUpdateApplicationComponentHandler creates a new handler
func NewUpdateApplicationComponentHandler(repository *repositories.ApplicationComponentRepository) *UpdateApplicationComponentHandler {
	return &UpdateApplicationComponentHandler{
		repository: repository,
	}
}

// Handle processes the command
func (h *UpdateApplicationComponentHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.UpdateApplicationComponent)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	// Load existing aggregate
	component, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	// Create value objects with validation
	name, err := valueobjects.NewComponentName(command.Name)
	if err != nil {
		return err
	}

	description := valueobjects.NewDescription(command.Description)

	// Update aggregate
	if err := component.Update(name, description); err != nil {
		return err
	}

	// Save to repository
	return h.repository.Save(ctx, component)
}
