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

func (h *RemoveStrategyPillarHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.RemoveStrategyPillar)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	config, err := h.repository.GetByID(ctx, command.ConfigID)
	if err != nil {
		return err
	}

	modifiedBy, err := valueobjects.NewUserEmail(command.ModifiedBy)
	if err != nil {
		return err
	}

	pillarID, err := valueobjects.NewStrategyPillarIDFromString(command.PillarID)
	if err != nil {
		return err
	}

	if err := config.RemoveStrategyPillar(pillarID, modifiedBy); err != nil {
		return err
	}

	return h.repository.Save(ctx, config)
}
