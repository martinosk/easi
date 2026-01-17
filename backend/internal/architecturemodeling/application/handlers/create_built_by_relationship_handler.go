package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type CreateBuiltByRelationshipHandler struct {
	repository *repositories.BuiltByRelationshipRepository
}

func NewCreateBuiltByRelationshipHandler(repository *repositories.BuiltByRelationshipRepository) *CreateBuiltByRelationshipHandler {
	return &CreateBuiltByRelationshipHandler{
		repository: repository,
	}
}

func (h *CreateBuiltByRelationshipHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CreateBuiltByRelationship)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	internalTeamID, err := valueobjects.NewInternalTeamIDFromString(command.InternalTeamID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	componentID, err := valueobjects.NewComponentIDFromString(command.ComponentID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	notes, err := valueobjects.NewNotes(command.Notes)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	relationship, err := aggregates.NewBuiltByRelationship(internalTeamID, componentID, notes)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, relationship); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(relationship.ID()), nil
}
