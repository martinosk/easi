package handlers

import (
	"context"
	"time"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

type AddCapabilityExpertRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.Capability, error)
	Save(ctx context.Context, capability *aggregates.Capability) error
}

type AddCapabilityExpertHandler struct {
	repository AddCapabilityExpertRepository
}

func NewAddCapabilityExpertHandler(repository AddCapabilityExpertRepository) *AddCapabilityExpertHandler {
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

	expert, err := valueobjects.NewExpert(command.ExpertName, command.ExpertRole, command.ContactInfo, time.Now().UTC())
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
