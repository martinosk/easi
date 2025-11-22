package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

var ErrCapabilityHasChildren = errors.New("cannot delete capability with children. Delete child capabilities first")

type DeleteCapabilityHandler struct {
	repository *repositories.CapabilityRepository
	readModel  *readmodels.CapabilityReadModel
}

func NewDeleteCapabilityHandler(
	repository *repositories.CapabilityRepository,
	readModel *readmodels.CapabilityReadModel,
) *DeleteCapabilityHandler {
	return &DeleteCapabilityHandler{
		repository: repository,
		readModel:  readModel,
	}
}

func (h *DeleteCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.DeleteCapability)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	children, err := h.readModel.GetChildren(ctx, command.ID)
	if err != nil {
		return err
	}

	if len(children) > 0 {
		return ErrCapabilityHasChildren
	}

	capability, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	if err := capability.Delete(); err != nil {
		return err
	}

	return h.repository.Save(ctx, capability)
}
