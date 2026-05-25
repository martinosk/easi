package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

type CreateInternalTeamRepository interface {
	Save(ctx context.Context, team *aggregates.InternalTeam) error
}

type CreateInternalTeamHandler struct {
	repository CreateInternalTeamRepository
}

func NewCreateInternalTeamHandler(repository CreateInternalTeamRepository) *CreateInternalTeamHandler {
	return &CreateInternalTeamHandler{
		repository: repository,
	}
}

func (h *CreateInternalTeamHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CreateInternalTeam)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	name, err := valueobjects.NewEntityName(command.Name)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	notes, err := valueobjects.NewNotes(command.Notes)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	team, err := aggregates.NewInternalTeam(name, command.Department, command.ContactPerson, notes)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, team); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(team.ID()), nil
}
