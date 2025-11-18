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

func (h *UpdateViewEdgeTypeHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.UpdateViewEdgeType)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	_, err := valueobjects.NewEdgeType(command.EdgeType)
	if err != nil {
		return err
	}

	return h.layoutRepository.UpdateEdgeType(ctx, command.ViewID, command.EdgeType)
}
