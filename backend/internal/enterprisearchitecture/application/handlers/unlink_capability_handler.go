package handlers

import (
	"context"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/shared/cqrs"
)

type UnlinkRepository interface {
	Save(ctx context.Context, link *aggregates.EnterpriseCapabilityLink) error
	GetByID(ctx context.Context, id string) (*aggregates.EnterpriseCapabilityLink, error)
}

type UnlinkCapabilityHandler struct {
	repository UnlinkRepository
}

func NewUnlinkCapabilityHandler(
	repository UnlinkRepository,
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
