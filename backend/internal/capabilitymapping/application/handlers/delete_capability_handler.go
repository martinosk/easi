package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/shared/cqrs"
)

var ErrCapabilityHasChildren = errors.New("cannot delete capability with children. Delete child capabilities first")

type DeleteCapabilityRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.Capability, error)
	Save(ctx context.Context, capability *aggregates.Capability) error
}

type DeleteCapabilityReadModel interface {
	GetChildren(ctx context.Context, parentID string) ([]readmodels.CapabilityDTO, error)
}

type DeleteCapabilityHandler struct {
	repository DeleteCapabilityRepository
	readModel  DeleteCapabilityReadModel
}

func NewDeleteCapabilityHandler(
	repository DeleteCapabilityRepository,
	readModel DeleteCapabilityReadModel,
) *DeleteCapabilityHandler {
	return &DeleteCapabilityHandler{
		repository: repository,
		readModel:  readModel,
	}
}

func (h *DeleteCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.DeleteCapability)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	children, err := h.readModel.GetChildren(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if len(children) > 0 {
		return cqrs.EmptyResult(), ErrCapabilityHasChildren
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
