package handlers

import (
	"context"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/application/readmodels"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type SetDefaultViewHandler struct {
	repository *repositories.ArchitectureViewRepository
	readModel  *readmodels.ArchitectureViewReadModel
}

func NewSetDefaultViewHandler(repository *repositories.ArchitectureViewRepository, readModel *readmodels.ArchitectureViewReadModel) *SetDefaultViewHandler {
	return &SetDefaultViewHandler{
		repository: repository,
		readModel:  readModel,
	}
}

func (h *SetDefaultViewHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.SetDefaultView)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	currentDefault, err := h.readModel.GetDefaultView(ctx)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if currentDefault != nil && currentDefault.ID != command.ViewID {
		oldDefaultView, err := h.repository.GetByID(ctx, currentDefault.ID)
		if err != nil {
			return cqrs.EmptyResult(), err
		}

		if err := oldDefaultView.UnsetAsDefault(); err != nil {
			return cqrs.EmptyResult(), err
		}

		if err := h.repository.Save(ctx, oldDefaultView); err != nil {
			return cqrs.EmptyResult(), err
		}
	}

	view, err := h.repository.GetByID(ctx, command.ViewID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := view.SetAsDefault(); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, view); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
