package handlers

import (
	"context"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/shared/cqrs"
)

type RemoveImportanceRepository interface {
	Save(ctx context.Context, importance *aggregates.EnterpriseStrategicImportance) error
	GetByID(ctx context.Context, id string) (*aggregates.EnterpriseStrategicImportance, error)
}

type RemoveEnterpriseStrategicImportanceHandler struct {
	repository RemoveImportanceRepository
}

func NewRemoveEnterpriseStrategicImportanceHandler(
	repository RemoveImportanceRepository,
) *RemoveEnterpriseStrategicImportanceHandler {
	return &RemoveEnterpriseStrategicImportanceHandler{
		repository: repository,
	}
}

func (h *RemoveEnterpriseStrategicImportanceHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.RemoveEnterpriseStrategicImportance)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	si, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := si.Remove(); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, si); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
