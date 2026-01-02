package handlers

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/services"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

type DeleteCapabilityRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.Capability, error)
	Save(ctx context.Context, capability *aggregates.Capability) error
}

type DeleteCapabilityHandler struct {
	repository      DeleteCapabilityRepository
	deletionService services.CapabilityDeletionService
}

func NewDeleteCapabilityHandler(
	repository DeleteCapabilityRepository,
	deletionService services.CapabilityDeletionService,
) *DeleteCapabilityHandler {
	return &DeleteCapabilityHandler{
		repository:      repository,
		deletionService: deletionService,
	}
}

func (h *DeleteCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.DeleteCapability)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	capabilityID, err := valueobjects.NewCapabilityIDFromString(command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.deletionService.CanDelete(ctx, capabilityID); err != nil {
		return cqrs.EmptyResult(), err
	}

	capability, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := capability.Delete(); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, capability); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
