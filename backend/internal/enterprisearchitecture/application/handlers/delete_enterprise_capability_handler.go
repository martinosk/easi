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

func (h *DeleteEnterpriseCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.DeleteEnterpriseCapability)
	if !ok {
		return cqrs.ErrInvalidCommand
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
