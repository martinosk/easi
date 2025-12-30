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

type SetEnterpriseStrategicImportanceHandler struct {
	repository          *repositories.EnterpriseStrategicImportanceRepository
	capabilityReadModel *readmodels.EnterpriseCapabilityReadModel
	importanceReadModel *readmodels.EnterpriseStrategicImportanceReadModel
}

func NewSetEnterpriseStrategicImportanceHandler(
	repository *repositories.EnterpriseStrategicImportanceRepository,
	capabilityReadModel *readmodels.EnterpriseCapabilityReadModel,
	importanceReadModel *readmodels.EnterpriseStrategicImportanceReadModel,
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
