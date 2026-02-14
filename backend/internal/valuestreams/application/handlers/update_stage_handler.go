package handlers

import (
	"context"

	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/valuestreams/application/commands"
	"easi/backend/internal/valuestreams/domain/aggregates"
	"easi/backend/internal/valuestreams/domain/valueobjects"
)

type UpdateStageRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.ValueStream, error)
	Save(ctx context.Context, vs *aggregates.ValueStream) error
}

type UpdateStageHandler struct {
	repository UpdateStageRepository
}

func NewUpdateStageHandler(repository UpdateStageRepository) *UpdateStageHandler {
	return &UpdateStageHandler{repository: repository}
}

func (h *UpdateStageHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.UpdateStage)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	vs, err := h.repository.GetByID(ctx, command.ValueStreamID)
	if err != nil {
		return cqrs.EmptyResult(), mapRepositoryError(err)
	}

	stageID, err := valueobjects.NewStageIDFromString(command.StageID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	name, err := valueobjects.NewStageName(command.Name)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	description, err := valueobjects.NewDescription(command.Description)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := vs.UpdateStage(stageID, name, description); err != nil {
		return cqrs.EmptyResult(), mapStageError(err)
	}

	if err := h.repository.Save(ctx, vs); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
