package handlers

import (
	"context"
	"time"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

type RemoveCapabilityExpertRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.Capability, error)
	Save(ctx context.Context, capability *aggregates.Capability) error
}

type RemoveCapabilityExpertHandler struct {
	repository RemoveCapabilityExpertRepository
}

func NewRemoveCapabilityExpertHandler(repository RemoveCapabilityExpertRepository) *RemoveCapabilityExpertHandler {
	return &RemoveCapabilityExpertHandler{
		repository: repository,
	}
}

func (h *RemoveCapabilityExpertHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.RemoveCapabilityExpert)
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

	if err := capability.RemoveExpert(expert); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, capability); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
