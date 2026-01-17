package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type CreateAcquiredViaRelationshipHandler struct {
	repository *repositories.AcquiredViaRelationshipRepository
}

func NewCreateAcquiredViaRelationshipHandler(repository *repositories.AcquiredViaRelationshipRepository) *CreateAcquiredViaRelationshipHandler {
	return &CreateAcquiredViaRelationshipHandler{
		repository: repository,
	}
}

func (h *CreateAcquiredViaRelationshipHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CreateAcquiredViaRelationship)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	acquiredEntityID, err := valueobjects.NewAcquiredEntityIDFromString(command.AcquiredEntityID)
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

	relationship, err := aggregates.NewAcquiredViaRelationship(acquiredEntityID, componentID, notes)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, relationship); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(relationship.ID()), nil
}
