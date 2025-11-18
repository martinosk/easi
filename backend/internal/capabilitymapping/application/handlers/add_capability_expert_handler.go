package handlers

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/entities"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type AddCapabilityExpertHandler struct {
	repository *repositories.CapabilityRepository
}

func NewAddCapabilityExpertHandler(repository *repositories.CapabilityRepository) *AddCapabilityExpertHandler {
	return &AddCapabilityExpertHandler{
		repository: repository,
	}
}

func (h *AddCapabilityExpertHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.AddCapabilityExpert)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	capability, err := h.repository.GetByID(ctx, command.CapabilityID)
	if err != nil {
		return err
	}

	expert, err := entities.NewExpert(command.ExpertName, command.ExpertRole, command.ContactInfo)
	if err != nil {
		return err
	}

	if err := capability.AddExpert(expert); err != nil {
		return err
	}

	return h.repository.Save(ctx, capability)
}
