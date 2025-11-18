package handlers

import (
	"context"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/domain/valueobjects"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type UpdateViewLayoutDirectionHandler struct {
	layoutRepository *repositories.ViewLayoutRepository
}

func NewUpdateViewLayoutDirectionHandler(layoutRepository *repositories.ViewLayoutRepository) *UpdateViewLayoutDirectionHandler {
	return &UpdateViewLayoutDirectionHandler{
		layoutRepository: layoutRepository,
	}
}

func (h *UpdateViewLayoutDirectionHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.UpdateViewLayoutDirection)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	_, err := valueobjects.NewLayoutDirection(command.LayoutDirection)
	if err != nil {
		return err
	}

	return h.layoutRepository.UpdateLayoutDirection(ctx, command.ViewID, command.LayoutDirection)
}
