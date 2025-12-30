package handlers

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type RemoveStrategyImportanceHandler struct {
	importanceRepo *repositories.StrategyImportanceRepository
}

func NewRemoveStrategyImportanceHandler(importanceRepo *repositories.StrategyImportanceRepository) *RemoveStrategyImportanceHandler {
	return &RemoveStrategyImportanceHandler{importanceRepo: importanceRepo}
}

func (h *RemoveStrategyImportanceHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.RemoveStrategyImportance)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	aggregate, err := h.importanceRepo.GetByID(ctx, command.ImportanceID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := aggregate.Remove(); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.importanceRepo.Save(ctx, aggregate); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
