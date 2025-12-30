package handlers

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type AddCapabilityTagHandler struct {
	repository *repositories.CapabilityRepository
}

func NewAddCapabilityTagHandler(repository *repositories.CapabilityRepository) *AddCapabilityTagHandler {
	return &AddCapabilityTagHandler{
		repository: repository,
	}
}

func (h *AddCapabilityTagHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.AddCapabilityTag)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	capability, err := h.repository.GetByID(ctx, command.CapabilityID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	tag, err := valueobjects.NewTag(command.Tag)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := capability.AddTag(tag); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, capability); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
