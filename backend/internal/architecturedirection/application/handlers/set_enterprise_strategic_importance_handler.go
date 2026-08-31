package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/application/readmodels"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

var ErrImportanceAlreadySet = errors.New("strategic importance already set for this pillar")

type SetImportanceRepository interface {
	Save(ctx context.Context, importance *aggregates.EnterpriseStrategicImportance) error
}

type SetImportanceCapabilityReadModel interface {
	GetByID(ctx context.Context, id string) (*readmodels.EnterpriseCapabilityDTO, error)
}

type SetImportanceReadModel interface {
	GetByCapabilityAndPillar(ctx context.Context, enterpriseCapabilityID, pillarID string) (*readmodels.EnterpriseStrategicImportanceDTO, error)
}

type SetEnterpriseStrategicImportanceHandler struct {
	repository          SetImportanceRepository
	capabilityReadModel SetImportanceCapabilityReadModel
	importanceReadModel SetImportanceReadModel
}

func NewSetEnterpriseStrategicImportanceHandler(
	repository SetImportanceRepository,
	capabilityReadModel SetImportanceCapabilityReadModel,
	importanceReadModel SetImportanceReadModel,
) *SetEnterpriseStrategicImportanceHandler {
	return &SetEnterpriseStrategicImportanceHandler{
		repository:          repository,
		capabilityReadModel: capabilityReadModel,
		importanceReadModel: importanceReadModel,
	}
}

func (h *SetEnterpriseStrategicImportanceHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.SetEnterpriseStrategicImportance)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	if err := h.rejectUnratablePillar(ctx, command); err != nil {
		return cqrs.EmptyResult(), err
	}

	params, err := newEnterpriseImportanceParams(command)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	si, err := aggregates.SetEnterpriseStrategicImportance(params)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, si); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(si.ID()), nil
}

func (h *SetEnterpriseStrategicImportanceHandler) rejectUnratablePillar(ctx context.Context, command *commands.SetEnterpriseStrategicImportance) error {
	capability, err := h.capabilityReadModel.GetByID(ctx, command.EnterpriseCapabilityID)
	if err != nil {
		return err
	}
	if capability == nil {
		return repositories.ErrEnterpriseCapabilityNotFound
	}

	rated, err := h.importanceReadModel.GetByCapabilityAndPillar(ctx, command.EnterpriseCapabilityID, command.PillarID)
	if err != nil {
		return err
	}
	if rated != nil {
		return ErrImportanceAlreadySet
	}

	return nil
}
