package handlers

import (
	"context"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

type UpdateImportanceRepository interface {
	Save(ctx context.Context, importance *aggregates.EnterpriseStrategicImportance) error
	GetByID(ctx context.Context, id string) (*aggregates.EnterpriseStrategicImportance, error)
}

type UpdateEnterpriseStrategicImportanceHandler struct {
	repository UpdateImportanceRepository
}

func NewUpdateEnterpriseStrategicImportanceHandler(
	repository UpdateImportanceRepository,
) *UpdateEnterpriseStrategicImportanceHandler {
	return &UpdateEnterpriseStrategicImportanceHandler{
		repository: repository,
	}
}

func (h *UpdateEnterpriseStrategicImportanceHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.UpdateEnterpriseStrategicImportance)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	si, err := h.repository.GetByID(ctx, command.ID)
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

	if err := si.Update(importance, rationale); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, si); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
