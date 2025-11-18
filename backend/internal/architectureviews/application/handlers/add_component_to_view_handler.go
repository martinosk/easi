package handlers

import (
	"context"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type AddComponentToViewHandler struct {
	viewRepository   *repositories.ArchitectureViewRepository
	layoutRepository *repositories.ViewLayoutRepository
}

func NewAddComponentToViewHandler(
	viewRepository *repositories.ArchitectureViewRepository,
	layoutRepository *repositories.ViewLayoutRepository,
) *AddComponentToViewHandler {
	return &AddComponentToViewHandler{
		viewRepository:   viewRepository,
		layoutRepository: layoutRepository,
	}
}

func (h *AddComponentToViewHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(commands.AddComponentToView)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	view, err := h.viewRepository.GetByID(ctx, command.ViewID)
	if err != nil {
		return err
	}

	if err := view.AddComponent(command.ComponentID); err != nil {
		return err
	}

	if err := h.viewRepository.Save(ctx, view); err != nil {
		return err
	}

	return h.layoutRepository.UpdateComponentPosition(ctx, command.ViewID, command.ComponentID, command.X, command.Y)
}
