package handlers

import (
	"context"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type UpdateMultiplePositionsHandler struct {
	layoutRepository *repositories.ViewLayoutRepository
}

func NewUpdateMultiplePositionsHandler(layoutRepository *repositories.ViewLayoutRepository) *UpdateMultiplePositionsHandler {
	return &UpdateMultiplePositionsHandler{
		layoutRepository: layoutRepository,
	}
}

func (h *UpdateMultiplePositionsHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(commands.UpdateMultiplePositions)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	positions := make([]repositories.ComponentPositionData, len(command.Positions))
	for i, pos := range command.Positions {
		positions[i] = repositories.ComponentPositionData{
			ComponentID: pos.ComponentID,
			X:           pos.X,
			Y:           pos.Y,
		}
	}

	if err := h.layoutRepository.UpdateMultiplePositions(ctx, command.ViewID, positions); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
