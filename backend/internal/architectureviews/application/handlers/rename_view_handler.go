package handlers

import (
	"context"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/domain/valueobjects"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type RenameViewHandler struct {
	repository *repositories.ArchitectureViewRepository
}

func NewRenameViewHandler(repository *repositories.ArchitectureViewRepository) *RenameViewHandler {
	return &RenameViewHandler{
		repository: repository,
	}
}

func (h *RenameViewHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.RenameView)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	view, err := h.repository.GetByID(ctx, command.ViewID)
	if err != nil {
		return err
	}

	newName, err := valueobjects.NewViewName(command.NewName)
	if err != nil {
		return err
	}

	if err := view.Rename(newName); err != nil {
		return err
	}

	return h.repository.Save(ctx, view)
}
