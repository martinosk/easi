package handlers

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type UpdateApplicationFitScoreHandler struct {
	repo *repositories.ApplicationFitScoreRepository
}

func NewUpdateApplicationFitScoreHandler(repo *repositories.ApplicationFitScoreRepository) *UpdateApplicationFitScoreHandler {
	return &UpdateApplicationFitScoreHandler{repo: repo}
}

func (h *UpdateApplicationFitScoreHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.UpdateApplicationFitScore)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	aggregate, err := h.repo.GetByID(ctx, command.FitScoreID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	score, err := valueobjects.NewFitScore(command.Score)
	if err != nil {
		return cqrs.EmptyResult(), ErrInvalidFitScoreValue
	}

	rationale, err := valueobjects.NewFitRationale(command.Rationale)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	updatedBy, err := valueobjects.NewUserIdentifier(command.UpdatedBy)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := aggregate.Update(score, rationale, updatedBy); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repo.Save(ctx, aggregate); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
