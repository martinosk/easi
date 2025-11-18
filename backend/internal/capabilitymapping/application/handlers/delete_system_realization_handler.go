package handlers

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type DeleteSystemRealizationHandler struct {
	repository *repositories.RealizationRepository
}

func NewDeleteSystemRealizationHandler(repository *repositories.RealizationRepository) *DeleteSystemRealizationHandler {
	return &DeleteSystemRealizationHandler{
		repository: repository,
	}
}

func (h *DeleteSystemRealizationHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.DeleteSystemRealization)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	realization, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	if err := realization.Delete(); err != nil {
		return err
	}

	return h.repository.Save(ctx, realization)
}
