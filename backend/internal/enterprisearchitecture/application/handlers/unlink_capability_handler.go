package handlers

import (
	"context"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type UnlinkCapabilityHandler struct {
	repository *repositories.EnterpriseCapabilityLinkRepository
}

func NewUnlinkCapabilityHandler(
	repository *repositories.EnterpriseCapabilityLinkRepository,
) *UnlinkCapabilityHandler {
	return &UnlinkCapabilityHandler{
		repository: repository,
	}
}

func (h *UnlinkCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.UnlinkCapability)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	link, err := h.repository.GetByID(ctx, command.LinkID)
	if err != nil {
		return err
	}

	if err := link.Unlink(); err != nil {
		return err
	}

	return h.repository.Save(ctx, link)
}
