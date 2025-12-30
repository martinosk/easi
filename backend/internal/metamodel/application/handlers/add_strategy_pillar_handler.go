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

func (h *AddStrategyPillarHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.AddStrategyPillar)
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

	pillarName, err := valueobjects.NewPillarName(command.Name)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	pillarDesc, err := valueobjects.NewPillarDescription(command.Description)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := config.AddStrategyPillar(pillarName, pillarDesc, modifiedBy); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, config); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
