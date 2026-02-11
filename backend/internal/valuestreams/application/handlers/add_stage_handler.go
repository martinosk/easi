package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/valuestreams/application/commands"
	"easi/backend/internal/valuestreams/domain/aggregates"
	"easi/backend/internal/valuestreams/domain/valueobjects"
	"easi/backend/internal/valuestreams/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

var ErrStageNameExists = errors.New("stage with this name already exists in this value stream")

type AddStageRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.ValueStream, error)
	Save(ctx context.Context, vs *aggregates.ValueStream) error
}

type AddStageHandler struct {
	repository AddStageRepository
}

func NewAddStageHandler(repository AddStageRepository) *AddStageHandler {
	return &AddStageHandler{repository: repository}
}

func (h *AddStageHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.AddStage)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	vs, err := h.repository.GetByID(ctx, command.ValueStreamID)
	if err != nil {
		if errors.Is(err, repositories.ErrValueStreamNotFound) {
			return cqrs.EmptyResult(), ErrValueStreamNotFound
		}
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

	var position *valueobjects.StagePosition
	if command.Position != nil {
		pos, err := valueobjects.NewStagePosition(*command.Position)
		if err != nil {
			return cqrs.EmptyResult(), err
		}
		position = &pos
	}

	stageID, err := vs.AddStage(name, description, position)
	if err != nil {
		if errors.Is(err, aggregates.ErrStageNameExists) {
			return cqrs.EmptyResult(), ErrStageNameExists
		}
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, vs); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(stageID.Value()), nil
}
