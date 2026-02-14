package handlers

import (
	"context"

	"easi/backend/internal/valuestreams/application/commands"
	"easi/backend/internal/valuestreams/domain/aggregates"
	"easi/backend/internal/shared/cqrs"
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
		positions[i] = aggregates.StagePositionUpdate{
			StageID:  p.StageID,
			Position: p.Position,
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
