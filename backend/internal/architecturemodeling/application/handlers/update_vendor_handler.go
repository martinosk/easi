package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

type UpdateVendorRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.Vendor, error)
	Save(ctx context.Context, vendor *aggregates.Vendor) error
}

type UpdateVendorHandler struct {
	repository UpdateVendorRepository
}

func NewUpdateVendorHandler(repository UpdateVendorRepository) *UpdateVendorHandler {
	return &UpdateVendorHandler{
		repository: repository,
	}
}

func (h *UpdateVendorHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.UpdateVendor)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	vendor, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	name, err := valueobjects.NewEntityName(command.Name)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	notes, err := valueobjects.NewNotes(command.Notes)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := vendor.Update(name, command.ImplementationPartner, notes); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, vendor); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
