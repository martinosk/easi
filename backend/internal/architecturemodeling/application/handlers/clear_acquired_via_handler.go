package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type ClearAcquiredViaHandler struct {
	repository *repositories.ComponentOriginsRepository
}

func NewClearAcquiredViaHandler(repository *repositories.ComponentOriginsRepository) *ClearAcquiredViaHandler {
	return &ClearAcquiredViaHandler{
		repository: repository,
	}
}

func (h *ClearAcquiredViaHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.ClearAcquiredVia)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	componentID, err := valueobjects.NewComponentIDFromString(command.ComponentID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	aggregateID := "component-origins:" + componentID.String()
	origins, err := h.repository.GetByID(ctx, aggregateID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := origins.ClearAcquiredVia(); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, origins); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(componentID.String()), nil
}
