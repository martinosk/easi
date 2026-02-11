package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/valuestreams/application/commands"
	"easi/backend/internal/valuestreams/domain/aggregates"
	"easi/backend/internal/valuestreams/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

var ErrValueStreamNameExists = errors.New("value stream with this name already exists")

type CreateValueStreamRepository interface {
	Save(ctx context.Context, vs *aggregates.ValueStream) error
}

type CreateValueStreamReadModel interface {
	NameExists(ctx context.Context, name, excludeID string) (bool, error)
}

type CreateValueStreamHandler struct {
	repository CreateValueStreamRepository
	readModel  CreateValueStreamReadModel
}

func NewCreateValueStreamHandler(
	repository CreateValueStreamRepository,
	readModel CreateValueStreamReadModel,
) *CreateValueStreamHandler {
	return &CreateValueStreamHandler{
		repository: repository,
		readModel:  readModel,
	}
}

func (h *CreateValueStreamHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CreateValueStream)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	exists, err := h.readModel.NameExists(ctx, command.Name, "")
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

	vs, err := aggregates.NewValueStream(name, description)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, vs); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(vs.ID()), nil
}
