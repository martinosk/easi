package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/shared/cqrs"
)

var ErrCapabilityHasLinks = errors.New("cannot delete enterprise capability: unlink all domain capabilities first")

type DeleteCapabilityRepository interface {
	Save(ctx context.Context, capability *aggregates.EnterpriseCapability) error
	GetByID(ctx context.Context, id string) (*aggregates.EnterpriseCapability, error)
}

type DeleteLinkCounter interface {
	CountByEnterpriseCapabilityID(ctx context.Context, enterpriseCapabilityID string) (int, error)
}

type DeleteEnterpriseCapabilityHandler struct {
	repository  DeleteCapabilityRepository
	linkCounter DeleteLinkCounter
}

func NewDeleteEnterpriseCapabilityHandler(
	repository DeleteCapabilityRepository,
	linkCounter DeleteLinkCounter,
) *DeleteEnterpriseCapabilityHandler {
	return &DeleteEnterpriseCapabilityHandler{
		repository:  repository,
		linkCounter: linkCounter,
	}
}

func (h *DeleteEnterpriseCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.DeleteEnterpriseCapability)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	linkCount, err := h.linkCounter.CountByEnterpriseCapabilityID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if linkCount > 0 {
		return cqrs.EmptyResult(), ErrCapabilityHasLinks
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
