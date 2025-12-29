package handlers

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type UpdateStrategyImportanceHandler struct {
	importanceRepo *repositories.StrategyImportanceRepository
}

func NewUpdateStrategyImportanceHandler(importanceRepo *repositories.StrategyImportanceRepository) *UpdateStrategyImportanceHandler {
	return &UpdateStrategyImportanceHandler{importanceRepo: importanceRepo}
}

func (h *UpdateStrategyImportanceHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.UpdateStrategyImportance)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	aggregate, err := h.importanceRepo.GetByID(ctx, command.ImportanceID)
	if err != nil {
		return err
	}

	importance, err := valueobjects.NewImportance(command.Importance)
	if err != nil {
		return ErrInvalidImportanceValue
	}

	rationale, err := valueobjects.NewRationale(command.Rationale)
	if err != nil {
		return err
	}

	if err := aggregate.Update(importance, rationale); err != nil {
		return err
	}

	return h.importanceRepo.Save(ctx, aggregate)
}
