package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/domain"
	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type CreateAcquiredViaRelationshipHandler struct {
	repository *repositories.AcquiredViaRelationshipRepository
	readModel  *readmodels.AcquiredViaRelationshipReadModel
}

func NewCreateAcquiredViaRelationshipHandler(
	repository *repositories.AcquiredViaRelationshipRepository,
	readModel *readmodels.AcquiredViaRelationshipReadModel,
) *CreateAcquiredViaRelationshipHandler {
	return &CreateAcquiredViaRelationshipHandler{
		repository: repository,
		readModel:  readModel,
	}
}

type acquiredViaParams struct {
	entityID    valueobjects.AcquiredEntityID
	componentID valueobjects.ComponentID
	notes       valueobjects.Notes
}

func parseAcquiredViaParams(cmd *commands.CreateAcquiredViaRelationship) (*acquiredViaParams, error) {
	entityID, err := valueobjects.NewAcquiredEntityIDFromString(cmd.AcquiredEntityID)
	if err != nil {
		return nil, err
	}

	componentID, err := valueobjects.NewComponentIDFromString(cmd.ComponentID)
	if err != nil {
		return nil, err
	}

	notes, err := valueobjects.NewNotes(cmd.Notes)
	if err != nil {
		return nil, err
	}

	return &acquiredViaParams{entityID: entityID, componentID: componentID, notes: notes}, nil
}

func (h *CreateAcquiredViaRelationshipHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CreateAcquiredViaRelationship)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	params, err := parseAcquiredViaParams(command)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.handleExistingRelationship(ctx, command); err != nil {
		return cqrs.EmptyResult(), err
	}

	relationship, err := aggregates.NewAcquiredViaRelationship(params.entityID, params.componentID, params.notes)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, relationship); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(relationship.ID()), nil
}

func (h *CreateAcquiredViaRelationshipHandler) handleExistingRelationship(ctx context.Context, cmd *commands.CreateAcquiredViaRelationship) error {
	existing, err := h.readModel.GetByComponentID(ctx, cmd.ComponentID)
	if err != nil {
		return err
	}

	if len(existing) == 0 {
		return nil
	}

	existingRel := existing[0]
	if !cmd.ReplaceExisting {
		return domain.NewRelationshipExistsError(
			existingRel.ID,
			existingRel.ComponentID,
			existingRel.AcquiredEntityID,
			existingRel.AcquiredEntityName,
			"AcquiredVia",
		)
	}

	return h.deleteExistingRelationship(ctx, existingRel.ID)
}

func (h *CreateAcquiredViaRelationshipHandler) deleteExistingRelationship(ctx context.Context, id string) error {
	existingAggregate, err := h.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := existingAggregate.Delete(); err != nil {
		return err
	}
	return h.repository.Save(ctx, existingAggregate)
}
