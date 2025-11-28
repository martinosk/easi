package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type CreateComponentRelationHandler struct {
	repository *repositories.ComponentRelationRepository
}

func NewCreateComponentRelationHandler(repository *repositories.ComponentRelationRepository) *CreateComponentRelationHandler {
	return &CreateComponentRelationHandler{
		repository: repository,
	}
}

func (h *CreateComponentRelationHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.CreateComponentRelation)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

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

	properties := valueobjects.NewRelationProperties(valueobjects.RelationPropertiesParams{
		SourceID:     sourceID,
		TargetID:     targetID,
		RelationType: relationType,
		Name:         name,
		Description:  description,
	})

	relation, err := aggregates.NewComponentRelation(properties)
	if err != nil {
		return err
	}

	command.ID = relation.ID()

	return h.repository.Save(ctx, relation)
}
