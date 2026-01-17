package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type DeletePurchasedFromRelationshipHandler struct {
	repository *repositories.PurchasedFromRelationshipRepository
}

func NewDeletePurchasedFromRelationshipHandler(repository *repositories.PurchasedFromRelationshipRepository) *DeletePurchasedFromRelationshipHandler {
	return &DeletePurchasedFromRelationshipHandler{
		repository: repository,
	}
}

func (h *DeletePurchasedFromRelationshipHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.DeletePurchasedFromRelationship)
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
