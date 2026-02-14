package handlers

import (
	"context"

	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/valuestreams/application/commands"
	"easi/backend/internal/valuestreams/domain/aggregates"
	"easi/backend/internal/valuestreams/domain/valueobjects"
)

type ReorderStagesRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.ValueStream, error)
	Save(ctx context.Context, vs *aggregates.ValueStream) error
}

type ReorderStagesHandler struct {
	repository ReorderStagesRepository
}

func NewReorderStagesHandler(repository ReorderStagesRepository) *ReorderStagesHandler {
	return &ReorderStagesHandler{repository: repository}
}

func (h *ReorderStagesHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.ReorderStages)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	vs, err := h.repository.GetByID(ctx, command.ValueStreamID)
	if err != nil {
		return cqrs.EmptyResult(), mapRepositoryError(err)
	}

	positions := make([]aggregates.StagePositionUpdate, len(command.Positions))
	for i, p := range command.Positions {
		stageID, err := valueobjects.NewStageIDFromString(p.StageID)
		if err != nil {
			return cqrs.EmptyResult(), err
		}
		pos, err := valueobjects.NewStagePosition(p.Position)
		if err != nil {
			return cqrs.EmptyResult(), err
		}
		positions[i] = aggregates.StagePositionUpdate{
			StageID:  stageID,
			Position: pos,
		}
	}

	if err := vs.ReorderStages(positions); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, vs); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
