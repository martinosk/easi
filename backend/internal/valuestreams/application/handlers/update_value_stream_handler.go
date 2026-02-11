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

var ErrValueStreamNotFound = errors.New("value stream not found")

type UpdateValueStreamRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.ValueStream, error)
	Save(ctx context.Context, vs *aggregates.ValueStream) error
}

type UpdateValueStreamReadModel interface {
	NameExists(ctx context.Context, name, excludeID string) (bool, error)
}

type UpdateValueStreamHandler struct {
	repository UpdateValueStreamRepository
	readModel  UpdateValueStreamReadModel
}

func NewUpdateValueStreamHandler(
	repository UpdateValueStreamRepository,
	readModel UpdateValueStreamReadModel,
) *UpdateValueStreamHandler {
	return &UpdateValueStreamHandler{
		repository: repository,
		readModel:  readModel,
	}
}

func (h *UpdateValueStreamHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.UpdateValueStream)
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

	exists, err := h.readModel.NameExists(ctx, command.Name, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if exists {
		return cqrs.EmptyResult(), ErrValueStreamNameExists
	}

	name, err := valueobjects.NewValueStreamName(command.Name)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	description, err := valueobjects.NewDescription(command.Description)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := vs.Update(name, description); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, vs); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
