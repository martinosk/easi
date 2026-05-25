package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

type UpdateComponentRelationRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.ComponentRelation, error)
	Save(ctx context.Context, relation *aggregates.ComponentRelation) error
}

type UpdateComponentRelationHandler struct {
	repository UpdateComponentRelationRepository
}

func NewUpdateComponentRelationHandler(repository UpdateComponentRelationRepository) *UpdateComponentRelationHandler {
	return &UpdateComponentRelationHandler{
		repository: repository,
	}
}

func (h *UpdateComponentRelationHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.UpdateComponentRelation)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	relation, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	name, err := valueobjects.NewDescription(command.Name)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	description, err := valueobjects.NewDescription(command.Description)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := relation.Update(name, description); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, relation); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
