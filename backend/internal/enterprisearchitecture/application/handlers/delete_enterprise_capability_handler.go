package handlers

import (
	"context"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/shared/cqrs"
)

type DeleteCapabilityRepository interface {
	Save(ctx context.Context, capability *aggregates.EnterpriseCapability) error
	GetByID(ctx context.Context, id string) (*aggregates.EnterpriseCapability, error)
}

type DeleteEnterpriseCapabilityHandler struct {
	repository DeleteCapabilityRepository
}

func NewDeleteEnterpriseCapabilityHandler(
	repository DeleteCapabilityRepository,
) *DeleteEnterpriseCapabilityHandler {
	return &DeleteEnterpriseCapabilityHandler{
		repository: repository,
	}
}

func (h *DeleteEnterpriseCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.DeleteEnterpriseCapability)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
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
