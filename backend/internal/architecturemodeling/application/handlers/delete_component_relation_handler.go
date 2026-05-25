package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

type DeleteComponentRelationRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.ComponentRelation, error)
	Save(ctx context.Context, relation *aggregates.ComponentRelation) error
}

type DeleteComponentRelationHandler struct {
	repository DeleteComponentRelationRepository
}

func NewDeleteComponentRelationHandler(repository DeleteComponentRelationRepository) *DeleteComponentRelationHandler {
	return &DeleteComponentRelationHandler{
		repository: repository,
	}
}

func (h *DeleteComponentRelationHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.DeleteComponentRelation)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	relationID, err := valueobjects.NewRelationIDFromString(command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	relation, err := h.repository.GetByID(ctx, relationID.Value())
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if relation.IsDeleted() {
		return cqrs.EmptyResult(), nil
	}

	if err := relation.Delete(); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, relation); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
