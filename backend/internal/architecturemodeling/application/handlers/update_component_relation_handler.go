package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

// UpdateComponentRelationHandler handles UpdateComponentRelation commands
type UpdateComponentRelationHandler struct {
	repository *repositories.ComponentRelationRepository
}

// NewUpdateComponentRelationHandler creates a new handler
func NewUpdateComponentRelationHandler(repository *repositories.ComponentRelationRepository) *UpdateComponentRelationHandler {
	return &UpdateComponentRelationHandler{
		repository: repository,
	}
}

// Handle processes the command
func (h *UpdateComponentRelationHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.UpdateComponentRelation)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	// Load existing aggregate
	relation, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	// Create value objects
	name := valueobjects.NewDescription(command.Name)
	description := valueobjects.NewDescription(command.Description)

	// Update aggregate
	if err := relation.Update(name, description); err != nil {
		return err
	}

	// Save to repository
	return h.repository.Save(ctx, relation)
}
