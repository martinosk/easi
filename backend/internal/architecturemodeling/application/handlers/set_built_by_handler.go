package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type SetBuiltByHandler struct {
	repository *repositories.ComponentOriginsRepository
}

func NewSetBuiltByHandler(repository *repositories.ComponentOriginsRepository) *SetBuiltByHandler {
	return &SetBuiltByHandler{
		repository: repository,
	}
}

func (h *SetBuiltByHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.SetBuiltBy)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	componentID, err := valueobjects.NewComponentIDFromString(command.ComponentID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	teamID, err := valueobjects.NewInternalTeamIDFromString(command.TeamID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	notes, err := valueobjects.NewNotes(command.Notes)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	origins, err := getOrCreateComponentOrigins(ctx, h.repository, componentID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := origins.SetBuiltBy(teamID, notes); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, origins); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(componentID.String()), nil
}
