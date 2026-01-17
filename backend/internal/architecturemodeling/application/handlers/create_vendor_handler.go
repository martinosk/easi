package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type CreateVendorHandler struct {
	repository *repositories.VendorRepository
}

func NewCreateVendorHandler(repository *repositories.VendorRepository) *CreateVendorHandler {
	return &CreateVendorHandler{
		repository: repository,
	}
}

func (h *CreateVendorHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CreateVendor)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	name, err := valueobjects.NewEntityName(command.Name)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	notes, err := valueobjects.NewNotes(command.Notes)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	vendor, err := aggregates.NewVendor(name, command.ImplementationPartner, notes)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, vendor); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(vendor.ID()), nil
}
