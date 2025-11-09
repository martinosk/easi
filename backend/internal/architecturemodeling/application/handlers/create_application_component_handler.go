package handlers

import (
	"context"

	"github.com/easi/backend/internal/architecturemodeling/application/commands"
	"github.com/easi/backend/internal/architecturemodeling/domain/aggregates"
	"github.com/easi/backend/internal/architecturemodeling/domain/valueobjects"
	"github.com/easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"github.com/easi/backend/internal/shared/cqrs"
)

// CreateApplicationComponentHandler handles CreateApplicationComponent commands
type CreateApplicationComponentHandler struct {
	repository *repositories.ApplicationComponentRepository
}

// NewCreateApplicationComponentHandler creates a new handler
func NewCreateApplicationComponentHandler(repository *repositories.ApplicationComponentRepository) *CreateApplicationComponentHandler {
	return &CreateApplicationComponentHandler{
		repository: repository,
	}
}

// Handle processes the command
func (h *CreateApplicationComponentHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.CreateApplicationComponent)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	// Create value objects with validation
	name, err := valueobjects.NewComponentName(command.Name)
	if err != nil {
		return err
	}

	description := valueobjects.NewDescription(command.Description)

	// Create aggregate
	component, err := aggregates.NewApplicationComponent(name, description)
	if err != nil {
		return err
	}

	// Set the ID on the command so the caller can retrieve it
	command.ID = component.ID()

	// Save to repository
	return h.repository.Save(ctx, component)
}
