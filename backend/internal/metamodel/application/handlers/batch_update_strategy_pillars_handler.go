package handlers

import (
	"context"

	"easi/backend/internal/metamodel/application/commands"
	"easi/backend/internal/metamodel/domain/valueobjects"
	"easi/backend/internal/metamodel/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
	domain "easi/backend/internal/shared/eventsourcing"
)

type BatchUpdateStrategyPillarsHandler struct {
	repository *repositories.MetaModelConfigurationRepository
}

func NewBatchUpdateStrategyPillarsHandler(repository *repositories.MetaModelConfigurationRepository) *BatchUpdateStrategyPillarsHandler {
	return &BatchUpdateStrategyPillarsHandler{
		repository: repository,
	}
}

func (h *BatchUpdateStrategyPillarsHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.BatchUpdateStrategyPillars)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	config, err := h.repository.GetByID(ctx, command.ConfigID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.validateAndApplyChanges(config, command); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, config); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}

func (h *BatchUpdateStrategyPillarsHandler) validateAndApplyChanges(config pillarConfig, command *commands.BatchUpdateStrategyPillars) error {
	if command.ExpectedVersion != nil && config.Version() != *command.ExpectedVersion {
		return domain.ErrConcurrencyConflict
	}

	modifiedBy, err := valueobjects.NewUserEmail(command.ModifiedBy)
	if err != nil {
		return err
	}

	for _, change := range command.Changes {
		if err := applyPillarChange(config, change, modifiedBy); err != nil {
			return err
		}
	}

	return nil
}

type pillarConfig interface {
	Version() int
	AddStrategyPillar(name valueobjects.PillarName, desc valueobjects.PillarDescription, modifiedBy valueobjects.UserEmail) error
	UpdateStrategyPillar(id valueobjects.StrategyPillarID, name valueobjects.PillarName, desc valueobjects.PillarDescription, modifiedBy valueobjects.UserEmail) error
	RemoveStrategyPillar(id valueobjects.StrategyPillarID, modifiedBy valueobjects.UserEmail) error
	UpdatePillarFitConfiguration(id valueobjects.StrategyPillarID, fitConfig valueobjects.FitConfigurationParams, modifiedBy valueobjects.UserEmail) error
}

func applyPillarChange(config pillarConfig, change commands.PillarChange, modifiedBy valueobjects.UserEmail) error {
	switch change.Operation {
	case commands.PillarOperationAdd:
		return addPillar(config, change, modifiedBy)
	case commands.PillarOperationUpdate:
		return updatePillar(config, change, modifiedBy)
	case commands.PillarOperationRemove:
		return removePillar(config, change, modifiedBy)
	}
	return nil
}

func addPillar(config pillarConfig, change commands.PillarChange, modifiedBy valueobjects.UserEmail) error {
	pillarName, err := valueobjects.NewPillarName(change.Name)
	if err != nil {
		return err
	}
	pillarDesc, err := valueobjects.NewPillarDescription(change.Description)
	if err != nil {
		return err
	}
	return config.AddStrategyPillar(pillarName, pillarDesc, modifiedBy)
}

func updatePillar(config pillarConfig, change commands.PillarChange, modifiedBy valueobjects.UserEmail) error {
	pillarID, pillarName, pillarDesc, err := parsePillarIdentity(change)
	if err != nil {
		return err
	}
	if err := config.UpdateStrategyPillar(pillarID, pillarName, pillarDesc, modifiedBy); err != nil {
		return err
	}
	return updateFitConfigIfEnabled(config, pillarID, change, modifiedBy)
}

func parsePillarIdentity(change commands.PillarChange) (valueobjects.StrategyPillarID, valueobjects.PillarName, valueobjects.PillarDescription, error) {
	pillarID, err := valueobjects.NewStrategyPillarIDFromString(change.PillarID)
	if err != nil {
		return valueobjects.StrategyPillarID{}, valueobjects.PillarName{}, valueobjects.PillarDescription{}, err
	}
	pillarName, err := valueobjects.NewPillarName(change.Name)
	if err != nil {
		return valueobjects.StrategyPillarID{}, valueobjects.PillarName{}, valueobjects.PillarDescription{}, err
	}
	pillarDesc, err := valueobjects.NewPillarDescription(change.Description)
	if err != nil {
		return valueobjects.StrategyPillarID{}, valueobjects.PillarName{}, valueobjects.PillarDescription{}, err
	}
	return pillarID, pillarName, pillarDesc, nil
}

func updateFitConfigIfEnabled(config pillarConfig, pillarID valueobjects.StrategyPillarID, change commands.PillarChange, modifiedBy valueobjects.UserEmail) error {
	if change.FitScoringEnabled == nil {
		return nil
	}
	criteria, err := valueobjects.NewFitCriteria(change.FitCriteria)
	if err != nil {
		return err
	}
	fitType, err := valueobjects.NewFitType(change.FitType)
	if err != nil {
		return err
	}
	fitConfig := valueobjects.NewFitConfigurationParams(*change.FitScoringEnabled, criteria, fitType)
	return config.UpdatePillarFitConfiguration(pillarID, fitConfig, modifiedBy)
}

func removePillar(config pillarConfig, change commands.PillarChange, modifiedBy valueobjects.UserEmail) error {
	pillarID, err := valueobjects.NewStrategyPillarIDFromString(change.PillarID)
	if err != nil {
		return err
	}
	return config.RemoveStrategyPillar(pillarID, modifiedBy)
}
