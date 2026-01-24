package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type SetPurchasedFromHandler struct {
	repository *repositories.ComponentOriginsRepository
}

func NewSetPurchasedFromHandler(repository *repositories.ComponentOriginsRepository) *SetPurchasedFromHandler {
	return &SetPurchasedFromHandler{
		repository: repository,
	}
}

func (h *SetPurchasedFromHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.SetPurchasedFrom)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	componentID, err := valueobjects.NewComponentIDFromString(command.ComponentID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	vendorID, err := valueobjects.NewVendorIDFromString(command.VendorID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	notes, err := valueobjects.NewNotes(command.Notes)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	origins, err := getOrCreateComponentOrigins(ctx, h.repository, componentID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := origins.SetPurchasedFrom(vendorID, notes); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, origins); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(componentID.String()), nil
}
