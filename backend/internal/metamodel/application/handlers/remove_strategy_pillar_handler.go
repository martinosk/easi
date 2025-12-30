package handlers

import (
	"context"

	"easi/backend/internal/metamodel/application/commands"
	"easi/backend/internal/metamodel/domain/valueobjects"
	"easi/backend/internal/metamodel/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type RemoveStrategyPillarHandler struct {
	repository *repositories.MetaModelConfigurationRepository
}

func NewRemoveStrategyPillarHandler(repository *repositories.MetaModelConfigurationRepository) *RemoveStrategyPillarHandler {
	return &RemoveStrategyPillarHandler{
		repository: repository,
	}
}

func (h *RemoveStrategyPillarHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.RemoveStrategyPillar)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	config, err := h.repository.GetByID(ctx, command.ConfigID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	modifiedBy, err := valueobjects.NewUserEmail(command.ModifiedBy)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	pillarID, err := valueobjects.NewStrategyPillarIDFromString(command.PillarID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := config.RemoveStrategyPillar(pillarID, modifiedBy); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, config); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
