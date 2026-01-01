package handlers

import (
	"context"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

type SetTargetMaturityRepository interface {
	Save(ctx context.Context, capability *aggregates.EnterpriseCapability) error
	GetByID(ctx context.Context, id string) (*aggregates.EnterpriseCapability, error)
}

type SetTargetMaturityHandler struct {
	repository SetTargetMaturityRepository
}

func NewSetTargetMaturityHandler(
	repository SetTargetMaturityRepository,
) *SetTargetMaturityHandler {
	return &SetTargetMaturityHandler{
		repository: repository,
	}
}

func (h *SetTargetMaturityHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.SetTargetMaturity)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	capability, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	targetMaturity, err := valueobjects.NewTargetMaturity(command.TargetMaturity)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := capability.SetTargetMaturity(targetMaturity); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, capability); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
