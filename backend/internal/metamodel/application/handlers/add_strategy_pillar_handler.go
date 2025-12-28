package handlers

import (
	"context"

	"easi/backend/internal/metamodel/application/commands"
	"easi/backend/internal/metamodel/domain/valueobjects"
	"easi/backend/internal/metamodel/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type AddStrategyPillarHandler struct {
	repository *repositories.MetaModelConfigurationRepository
}

func NewAddStrategyPillarHandler(repository *repositories.MetaModelConfigurationRepository) *AddStrategyPillarHandler {
	return &AddStrategyPillarHandler{
		repository: repository,
	}
}

func (h *AddStrategyPillarHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.AddStrategyPillar)
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

	pillarName, err := valueobjects.NewPillarName(command.Name)
	if err != nil {
		return err
	}

	pillarDesc, err := valueobjects.NewPillarDescription(command.Description)
	if err != nil {
		return err
	}

	if err := config.AddStrategyPillar(pillarName, pillarDesc, modifiedBy); err != nil {
		return err
	}

	return h.repository.Save(ctx, config)
}
