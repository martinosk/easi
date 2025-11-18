package handlers

import (
	"context"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type UpdateComponentPositionHandler struct {
	layoutRepository *repositories.ViewLayoutRepository
}

func NewUpdateComponentPositionHandler(layoutRepository *repositories.ViewLayoutRepository) *UpdateComponentPositionHandler {
	return &UpdateComponentPositionHandler{
		layoutRepository: layoutRepository,
	}
}

func (h *UpdateComponentPositionHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(commands.UpdateComponentPosition)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	return h.layoutRepository.UpdateComponentPosition(ctx, command.ViewID, command.ComponentID, command.X, command.Y)
}
