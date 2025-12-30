package handlers

import (
	"context"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/application/readmodels"
	"easi/backend/internal/architectureviews/domain/aggregates"
	"easi/backend/internal/architectureviews/domain/valueobjects"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type CreateViewHandler struct {
	repository *repositories.ArchitectureViewRepository
	readModel  *readmodels.ArchitectureViewReadModel
}

func NewCreateViewHandler(repository *repositories.ArchitectureViewRepository, readModel *readmodels.ArchitectureViewReadModel) *CreateViewHandler {
	return &CreateViewHandler{
		repository: repository,
		readModel:  readModel,
	}
}

func (h *CreateViewHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CreateView)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	name, err := valueobjects.NewViewName(command.Name)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	existingViews, err := h.readModel.GetAll(ctx)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	isDefault := len(existingViews) == 0

	view, err := aggregates.NewArchitectureView(name, command.Description, isDefault)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, view); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(view.ID()), nil
}
