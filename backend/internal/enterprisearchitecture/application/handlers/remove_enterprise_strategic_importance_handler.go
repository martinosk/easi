package handlers

import (
	"context"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type RemoveEnterpriseStrategicImportanceHandler struct {
	repository *repositories.EnterpriseStrategicImportanceRepository
}

func NewRemoveEnterpriseStrategicImportanceHandler(
	repository *repositories.EnterpriseStrategicImportanceRepository,
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
