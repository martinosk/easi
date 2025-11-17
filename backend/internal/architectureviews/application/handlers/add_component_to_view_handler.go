package handlers

import (
	"context"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/domain/valueobjects"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type AddComponentToViewHandler struct {
	repository *repositories.ArchitectureViewRepository
}

func NewAddComponentToViewHandler(repository *repositories.ArchitectureViewRepository) *AddComponentToViewHandler {
	return &AddComponentToViewHandler{
		repository: repository,
	}
}

func (h *AddComponentToViewHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(commands.AddComponentToView)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	view, err := h.repository.GetByID(ctx, command.ViewID)
	if err != nil {
		return err
	}

	position := valueobjects.NewComponentPosition(command.X, command.Y)

	if err := view.AddComponent(command.ComponentID, position); err != nil {
		return err
	}

	return h.repository.Save(ctx, view)
}
