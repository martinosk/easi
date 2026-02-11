package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/valuestreams/application/commands"
	"easi/backend/internal/valuestreams/domain/aggregates"
	"easi/backend/internal/valuestreams/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type DeleteValueStreamRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.ValueStream, error)
	Save(ctx context.Context, vs *aggregates.ValueStream) error
}

type DeleteValueStreamHandler struct {
	repository DeleteValueStreamRepository
}

func NewDeleteValueStreamHandler(
	repository DeleteValueStreamRepository,
) *DeleteValueStreamHandler {
	return &DeleteValueStreamHandler{
		repository: repository,
	}
}

func (h *DeleteValueStreamHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.DeleteValueStream)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	vs, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrValueStreamNotFound) {
			return cqrs.EmptyResult(), ErrValueStreamNotFound
		}
		return cqrs.EmptyResult(), err
	}

	if err := vs.Delete(); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, vs); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
