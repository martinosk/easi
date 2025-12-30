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

func (h *UnlinkCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.UnlinkCapability)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	link, err := h.repository.GetByID(ctx, command.LinkID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := link.Unlink(); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, link); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
