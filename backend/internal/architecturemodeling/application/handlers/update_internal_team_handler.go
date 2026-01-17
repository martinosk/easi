package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type UpdateInternalTeamHandler struct {
	repository *repositories.InternalTeamRepository
}

func NewUpdateInternalTeamHandler(repository *repositories.InternalTeamRepository) *UpdateInternalTeamHandler {
	return &UpdateInternalTeamHandler{
		repository: repository,
	}
}

func (h *UpdateInternalTeamHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.UpdateInternalTeam)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	team, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	name, err := valueobjects.NewEntityName(command.Name)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	notes, err := valueobjects.NewNotes(command.Notes)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := team.Update(name, command.Department, command.ContactPerson, notes); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, team); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
