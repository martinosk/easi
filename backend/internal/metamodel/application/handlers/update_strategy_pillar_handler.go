package handlers

import (
	"context"

	"easi/backend/internal/metamodel/application/commands"
	"easi/backend/internal/metamodel/domain/aggregates"
	"easi/backend/internal/metamodel/domain/valueobjects"
	"easi/backend/internal/metamodel/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
	domain "easi/backend/internal/shared/eventsourcing"
)

type UpdateStrategyPillarHandler struct {
	repository *repositories.MetaModelConfigurationRepository
}

func NewUpdateStrategyPillarHandler(repository *repositories.MetaModelConfigurationRepository) *UpdateStrategyPillarHandler {
	return &UpdateStrategyPillarHandler{
		repository: repository,
	}
}

func (h *UpdateStrategyPillarHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.UpdateStrategyPillar)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	config, err := h.repository.GetByID(ctx, command.ConfigID)
	if err != nil {
		return err
	}

	if command.ExpectedVersion != nil && config.Version() != *command.ExpectedVersion {
		return domain.ErrConcurrencyConflict
	}

	if err := h.executeUpdate(config, command); err != nil {
		return err
	}

	return h.repository.Save(ctx, config)
}

func (h *UpdateStrategyPillarHandler) executeUpdate(config *aggregates.MetaModelConfiguration, command *commands.UpdateStrategyPillar) error {
	modifiedBy, err := valueobjects.NewUserEmail(command.ModifiedBy)
	if err != nil {
		return err
	}

	pillarID, err := valueobjects.NewStrategyPillarIDFromString(command.PillarID)
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

	return config.UpdateStrategyPillar(pillarID, pillarName, pillarDesc, modifiedBy)
}
