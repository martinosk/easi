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

func (h *UpdateComponentPositionHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(commands.UpdateComponentPosition)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	if err := h.layoutRepository.UpdateComponentPosition(ctx, command.ViewID, command.ComponentID, command.X, command.Y); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
