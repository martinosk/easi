package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type SetAcquiredViaHandler struct {
	repository *repositories.ComponentOriginsRepository
}

func NewSetAcquiredViaHandler(repository *repositories.ComponentOriginsRepository) *SetAcquiredViaHandler {
	return &SetAcquiredViaHandler{
		repository: repository,
	}
}

func (h *SetAcquiredViaHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.SetAcquiredVia)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	componentID, err := valueobjects.NewComponentIDFromString(command.ComponentID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	entityID, err := valueobjects.NewAcquiredEntityIDFromString(command.EntityID)
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

	if err := origins.SetAcquiredVia(entityID, notes); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, origins); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(componentID.String()), nil
}
