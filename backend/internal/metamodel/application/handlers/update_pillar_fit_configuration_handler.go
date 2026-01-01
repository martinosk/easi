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

type UpdatePillarFitConfigurationHandler struct {
	repository *repositories.MetaModelConfigurationRepository
}

func NewUpdatePillarFitConfigurationHandler(repository *repositories.MetaModelConfigurationRepository) *UpdatePillarFitConfigurationHandler {
	return &UpdatePillarFitConfigurationHandler{
		repository: repository,
	}
}

func (h *UpdatePillarFitConfigurationHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.UpdatePillarFitConfiguration)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	config, err := h.repository.GetByID(ctx, command.ConfigID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if command.ExpectedVersion != nil && config.Version() != *command.ExpectedVersion {
		return cqrs.EmptyResult(), domain.ErrConcurrencyConflict
	}

	if err := h.executeUpdate(config, command); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, config); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}

func (h *UpdatePillarFitConfigurationHandler) executeUpdate(config *aggregates.MetaModelConfiguration, command *commands.UpdatePillarFitConfiguration) error {
	modifiedBy, err := valueobjects.NewUserEmail(command.ModifiedBy)
	if err != nil {
		return err
	}

	pillarID, err := valueobjects.NewStrategyPillarIDFromString(command.PillarID)
	if err != nil {
		return err
	}

	criteria, err := valueobjects.NewFitCriteria(command.FitCriteria)
	if err != nil {
		return err
	}

	return config.UpdatePillarFitConfiguration(pillarID, command.FitScoringEnabled, criteria, modifiedBy)
}
