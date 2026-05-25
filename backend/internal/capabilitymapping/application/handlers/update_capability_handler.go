package handlers

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

type UpdateCapabilityRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.Capability, error)
	Save(ctx context.Context, capability *aggregates.Capability) error
}

type UpdateCapabilityHandler struct {
	repository UpdateCapabilityRepository
}

func NewUpdateCapabilityHandler(repository UpdateCapabilityRepository) *UpdateCapabilityHandler {
	return &UpdateCapabilityHandler{
		repository: repository,
	}
}

func (h *UpdateCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.UpdateCapability)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	capability, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	name, err := valueobjects.NewCapabilityName(command.Name)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	description, err := valueobjects.NewDescription(command.Description)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := capability.Update(name, description); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, capability); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
