package handlers

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type UpdateSystemRealizationHandler struct {
	repository *repositories.RealizationRepository
}

func NewUpdateSystemRealizationHandler(repository *repositories.RealizationRepository) *UpdateSystemRealizationHandler {
	return &UpdateSystemRealizationHandler{
		repository: repository,
	}
}

func (h *UpdateSystemRealizationHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.UpdateSystemRealization)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	realization, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	realizationLevel, err := valueobjects.NewRealizationLevel(command.RealizationLevel)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	notes, err := valueobjects.NewDescription(command.Notes)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := realization.Update(realizationLevel, notes); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, realization); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
