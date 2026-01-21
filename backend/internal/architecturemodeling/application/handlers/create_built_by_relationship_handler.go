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

type CreateBuiltByRelationshipHandler struct {
	repository *repositories.BuiltByRelationshipRepository
	readModel  *readmodels.BuiltByRelationshipReadModel
}

func NewCreateBuiltByRelationshipHandler(
	repository *repositories.BuiltByRelationshipRepository,
	readModel *readmodels.BuiltByRelationshipReadModel,
) *CreateBuiltByRelationshipHandler {
	return &CreateBuiltByRelationshipHandler{
		repository: repository,
		readModel:  readModel,
	}
}

type builtByParams struct {
	internalTeamID valueobjects.InternalTeamID
	componentID    valueobjects.ComponentID
	notes          valueobjects.Notes
}

func parseBuiltByParams(cmd *commands.CreateBuiltByRelationship) (*builtByParams, error) {
	internalTeamID, err := valueobjects.NewInternalTeamIDFromString(cmd.InternalTeamID)
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

	return &builtByParams{internalTeamID: internalTeamID, componentID: componentID, notes: notes}, nil
}

func (h *CreateBuiltByRelationshipHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CreateBuiltByRelationship)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	params, err := parseBuiltByParams(command)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.handleExistingRelationship(ctx, command); err != nil {
		return cqrs.EmptyResult(), err
	}

	relationship, err := aggregates.NewBuiltByRelationship(params.internalTeamID, params.componentID, params.notes)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, relationship); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(relationship.ID()), nil
}

func (h *CreateBuiltByRelationshipHandler) handleExistingRelationship(ctx context.Context, cmd *commands.CreateBuiltByRelationship) error {
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
			existingRel.InternalTeamID,
			existingRel.InternalTeamName,
			"BuiltBy",
		)
	}

	return h.deleteExistingRelationship(ctx, existingRel.ID)
}

func (h *CreateBuiltByRelationshipHandler) deleteExistingRelationship(ctx context.Context, id string) error {
	existingAggregate, err := h.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := existingAggregate.Delete(); err != nil {
		return err
	}
	return h.repository.Save(ctx, existingAggregate)
}
