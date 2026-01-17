package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type CreatePurchasedFromRelationshipHandler struct {
	repository *repositories.PurchasedFromRelationshipRepository
}

func NewCreatePurchasedFromRelationshipHandler(repository *repositories.PurchasedFromRelationshipRepository) *CreatePurchasedFromRelationshipHandler {
	return &CreatePurchasedFromRelationshipHandler{
		repository: repository,
	}
}

func (h *CreatePurchasedFromRelationshipHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CreatePurchasedFromRelationship)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	vendorID, err := valueobjects.NewVendorIDFromString(command.VendorID)
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

	relationship, err := aggregates.NewPurchasedFromRelationship(vendorID, componentID, notes)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, relationship); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(relationship.ID()), nil
}
