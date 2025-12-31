package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"
	"easi/backend/internal/enterprisearchitecture/infrastructure/repositories"
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

	capability, err := h.capabilityReadModel.GetByID(ctx, command.EnterpriseCapabilityID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if capability == nil {
		return cqrs.EmptyResult(), repositories.ErrEnterpriseCapabilityNotFound
	}

	existing, err := h.importanceReadModel.GetByCapabilityAndPillar(ctx, command.EnterpriseCapabilityID, command.PillarID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if existing != nil {
		return cqrs.EmptyResult(), ErrImportanceAlreadySet
	}

	enterpriseCapabilityID, err := valueobjects.NewEnterpriseCapabilityIDFromString(command.EnterpriseCapabilityID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	pillarID, err := valueobjects.NewPillarIDFromString(command.PillarID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	importance, err := valueobjects.NewImportance(command.Importance)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	rationale, err := valueobjects.NewRationale(command.Rationale)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	si, err := aggregates.SetEnterpriseStrategicImportance(aggregates.NewEnterpriseImportanceParams{
		EnterpriseCapabilityID: enterpriseCapabilityID,
		PillarID:               pillarID,
		PillarName:             command.PillarName,
		Importance:             importance,
		Rationale:              rationale,
	})
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, si); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(si.ID()), nil
}
