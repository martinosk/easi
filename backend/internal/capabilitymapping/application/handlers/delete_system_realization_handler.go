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

func NewDeleteSystemRealizationHandler(
	repository *repositories.RealizationRepository,
) *DeleteSystemRealizationHandler {
	return &DeleteSystemRealizationHandler{
		repository: repository,
	}
}

func (h *DeleteSystemRealizationHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.DeleteSystemRealization)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	realization, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := realization.Delete(); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, realization); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
