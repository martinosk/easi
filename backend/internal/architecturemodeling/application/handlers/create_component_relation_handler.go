package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

// CreateComponentRelationHandler handles CreateComponentRelation commands
type CreateComponentRelationHandler struct {
	repository *repositories.ComponentRelationRepository
}

// NewCreateComponentRelationHandler creates a new handler
func NewCreateComponentRelationHandler(repository *repositories.ComponentRelationRepository) *CreateComponentRelationHandler {
	return &CreateComponentRelationHandler{
		repository: repository,
	}
}

// Handle processes the command
func (h *CreateComponentRelationHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.CreateComponentRelation)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	// Create value objects with validation
	sourceID, err := valueobjects.NewComponentIDFromString(command.SourceComponentID)
	if err != nil {
		return err
	}

	targetID, err := valueobjects.NewComponentIDFromString(command.TargetComponentID)
	if err != nil {
		return err
	}

	relationType, err := valueobjects.NewRelationType(command.RelationType)
	if err != nil {
		return err
	}

	name := valueobjects.NewDescription(command.Name)
	description := valueobjects.NewDescription(command.Description)

	// Create aggregate
	relation, err := aggregates.NewComponentRelation(sourceID, targetID, relationType, name, description)
	if err != nil {
		return err
	}

	// Set the ID on the command so the caller can retrieve it
	command.ID = relation.ID()

	// Save to repository
	return h.repository.Save(ctx, relation)
}
