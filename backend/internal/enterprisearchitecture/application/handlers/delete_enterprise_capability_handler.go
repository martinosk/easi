package handlers

import (
	"context"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type DeleteEnterpriseCapabilityHandler struct {
	repository *repositories.EnterpriseCapabilityRepository
}

func NewDeleteEnterpriseCapabilityHandler(
	repository *repositories.EnterpriseCapabilityRepository,
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
