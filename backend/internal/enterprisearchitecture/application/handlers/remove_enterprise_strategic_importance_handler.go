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

func (h *RemoveEnterpriseStrategicImportanceHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.RemoveEnterpriseStrategicImportance)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	si, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	if err := si.Remove(); err != nil {
		return err
	}

	return h.repository.Save(ctx, si)
}
