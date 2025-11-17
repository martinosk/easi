package handlers

import (
	"context"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/domain/valueobjects"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type UpdateComponentPositionHandler struct {
	repository *repositories.ArchitectureViewRepository
}

func NewUpdateComponentPositionHandler(repository *repositories.ArchitectureViewRepository) *UpdateComponentPositionHandler {
	return &UpdateComponentPositionHandler{
		repository: repository,
	}
}

func (h *UpdateComponentPositionHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(commands.UpdateComponentPosition)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	view, err := h.repository.GetByID(ctx, command.ViewID)
	if err != nil {
		return err
	}

	newPosition := valueobjects.NewComponentPosition(command.X, command.Y)

	if err := view.UpdateComponentPosition(command.ComponentID, newPosition); err != nil {
		return err
	}

	return h.repository.Save(ctx, view)
}
