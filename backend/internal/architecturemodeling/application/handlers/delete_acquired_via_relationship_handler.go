package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type DeleteAcquiredViaRelationshipHandler struct {
	repository *repositories.AcquiredViaRelationshipRepository
}

func NewDeleteAcquiredViaRelationshipHandler(repository *repositories.AcquiredViaRelationshipRepository) *DeleteAcquiredViaRelationshipHandler {
	return &DeleteAcquiredViaRelationshipHandler{
		repository: repository,
	}
}

func (h *DeleteAcquiredViaRelationshipHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.DeleteAcquiredViaRelationship)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	relationship, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if relationship.IsDeleted() {
		return cqrs.EmptyResult(), nil
	}

	if err := relationship.Delete(); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, relationship); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
