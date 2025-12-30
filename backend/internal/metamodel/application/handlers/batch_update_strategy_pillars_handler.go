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

	if command.ExpectedVersion != nil && config.Version() != *command.ExpectedVersion {
		return cqrs.EmptyResult(), domain.ErrConcurrencyConflict
	}

	modifiedBy, err := valueobjects.NewUserEmail(command.ModifiedBy)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	for _, change := range command.Changes {
		if err := applyPillarChange(config, change, modifiedBy); err != nil {
			return cqrs.EmptyResult(), err
		}
	}

	if err := h.repository.Save(ctx, config); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}

type pillarConfig interface {
	AddStrategyPillar(name valueobjects.PillarName, desc valueobjects.PillarDescription, modifiedBy valueobjects.UserEmail) error
	UpdateStrategyPillar(id valueobjects.StrategyPillarID, name valueobjects.PillarName, desc valueobjects.PillarDescription, modifiedBy valueobjects.UserEmail) error
	RemoveStrategyPillar(id valueobjects.StrategyPillarID, modifiedBy valueobjects.UserEmail) error
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
	pillarID, err := valueobjects.NewStrategyPillarIDFromString(change.PillarID)
	if err != nil {
		return err
	}
	pillarName, err := valueobjects.NewPillarName(change.Name)
	if err != nil {
		return err
	}
	pillarDesc, err := valueobjects.NewPillarDescription(change.Description)
	if err != nil {
		return err
	}
	return config.UpdateStrategyPillar(pillarID, pillarName, pillarDesc, modifiedBy)
}

func removePillar(config pillarConfig, change commands.PillarChange, modifiedBy valueobjects.UserEmail) error {
	pillarID, err := valueobjects.NewStrategyPillarIDFromString(change.PillarID)
	if err != nil {
		return err
	}
	return config.RemoveStrategyPillar(pillarID, modifiedBy)
}
