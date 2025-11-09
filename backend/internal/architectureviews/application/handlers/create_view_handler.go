package handlers

import (
	"context"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/domain/aggregates"
	"easi/backend/internal/architectureviews/domain/valueobjects"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

// CreateViewHandler handles CreateView commands
type CreateViewHandler struct {
	repository *repositories.ArchitectureViewRepository
}

// NewCreateViewHandler creates a new handler
func NewCreateViewHandler(repository *repositories.ArchitectureViewRepository) *CreateViewHandler {
	return &CreateViewHandler{
		repository: repository,
	}
}

// Handle processes the command
func (h *CreateViewHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.CreateView)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	// Create value objects with validation
	name, err := valueobjects.NewViewName(command.Name)
	if err != nil {
		return err
	}

	// Create aggregate
	view, err := aggregates.NewArchitectureView(name, command.Description)
	if err != nil {
		return err
	}

	// Set the ID on the command so the caller can retrieve it
	command.ID = view.ID()

	// Save to repository
	return h.repository.Save(ctx, view)
}
