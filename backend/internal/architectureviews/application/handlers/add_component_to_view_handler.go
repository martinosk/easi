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

func (h *AddComponentToViewHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(commands.AddComponentToView)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	err := h.layoutRepository.RunInTx(ctx, func(ctx context.Context) error {
		return h.addAndPlace(ctx, command)
	})
	return cqrs.EmptyResult(), err
}

func (h *AddComponentToViewHandler) addAndPlace(ctx context.Context, command commands.AddComponentToView) error {
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

	ref := repositories.ElementRef{ViewID: command.ViewID, ElementID: command.ComponentID, ElementType: repositories.ElementTypeComponent}
	return h.layoutRepository.MoveElement(ctx, ref, repositories.Position{X: command.X, Y: command.Y})
}
