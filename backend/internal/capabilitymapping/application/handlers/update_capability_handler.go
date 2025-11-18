package handlers

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type UpdateCapabilityHandler struct {
	repository *repositories.CapabilityRepository
}

func NewUpdateCapabilityHandler(repository *repositories.CapabilityRepository) *UpdateCapabilityHandler {
	return &UpdateCapabilityHandler{
		repository: repository,
	}
}

func (h *UpdateCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.UpdateCapability)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	capability, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	name, err := valueobjects.NewCapabilityName(command.Name)
	if err != nil {
		return err
	}

	description := valueobjects.NewDescription(command.Description)

	if err := capability.Update(name, description); err != nil {
		return err
	}

	return h.repository.Save(ctx, capability)
}
