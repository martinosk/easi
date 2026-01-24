package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type ClearBuiltByHandler struct {
	repository *repositories.ComponentOriginsRepository
}

func NewClearBuiltByHandler(repository *repositories.ComponentOriginsRepository) *ClearBuiltByHandler {
	return &ClearBuiltByHandler{
		repository: repository,
	}
}

func (h *ClearBuiltByHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.ClearBuiltBy)
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

	if err := origins.ClearBuiltBy(); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, origins); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(componentID.String()), nil
}
