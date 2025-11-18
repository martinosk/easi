package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type DeleteApplicationComponentHandler struct {
	repository *repositories.ApplicationComponentRepository
}

func NewDeleteApplicationComponentHandler(repository *repositories.ApplicationComponentRepository) *DeleteApplicationComponentHandler {
	return &DeleteApplicationComponentHandler{
		repository: repository,
	}
}

func (h *DeleteApplicationComponentHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.DeleteApplicationComponent)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	componentID, err := valueobjects.NewComponentIDFromString(command.ID)
	if err != nil {
		return err
	}

	component, err := h.repository.GetByID(ctx, componentID.Value())
	if err != nil {
		return err
	}

	if component.IsDeleted() {
		return nil
	}

	if err := component.Delete(); err != nil {
		return err
	}

	return h.repository.Save(ctx, component)
}
