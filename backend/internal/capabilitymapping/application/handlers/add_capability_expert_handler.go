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

func (h *AddCapabilityExpertHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.AddCapabilityExpert)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	capability, err := h.repository.GetByID(ctx, command.CapabilityID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	expert, err := entities.NewExpert(command.ExpertName, command.ExpertRole, command.ContactInfo)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := capability.AddExpert(expert); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, capability); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
