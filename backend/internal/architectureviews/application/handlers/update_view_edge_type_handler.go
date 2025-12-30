package handlers

import (
	"context"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/domain/valueobjects"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type UpdateViewEdgeTypeHandler struct {
	layoutRepository *repositories.ViewLayoutRepository
}

func NewUpdateViewEdgeTypeHandler(layoutRepository *repositories.ViewLayoutRepository) *UpdateViewEdgeTypeHandler {
	return &UpdateViewEdgeTypeHandler{
		layoutRepository: layoutRepository,
	}
}

func (h *UpdateViewEdgeTypeHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.UpdateViewEdgeType)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	_, err := valueobjects.NewEdgeType(command.EdgeType)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.layoutRepository.UpdateEdgeType(ctx, command.ViewID, command.EdgeType); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
