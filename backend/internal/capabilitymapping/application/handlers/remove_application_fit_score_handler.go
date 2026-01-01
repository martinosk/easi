package handlers

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type RemoveApplicationFitScoreHandler struct {
	repo *repositories.ApplicationFitScoreRepository
}

func NewRemoveApplicationFitScoreHandler(repo *repositories.ApplicationFitScoreRepository) *RemoveApplicationFitScoreHandler {
	return &RemoveApplicationFitScoreHandler{repo: repo}
}

func (h *RemoveApplicationFitScoreHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.RemoveApplicationFitScore)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	aggregate, err := h.repo.GetByID(ctx, command.FitScoreID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	removedBy, err := valueobjects.NewUserIdentifier(command.RemovedBy)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := aggregate.Remove(removedBy); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repo.Save(ctx, aggregate); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
