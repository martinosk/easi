package handlers

import (
	"context"

	"easi/backend/internal/metamodel/application/commands"
	"easi/backend/internal/metamodel/domain/valueobjects"
	"easi/backend/internal/metamodel/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type ResetMaturityScaleHandler struct {
	repository *repositories.MetaModelConfigurationRepository
}

func NewResetMaturityScaleHandler(repository *repositories.MetaModelConfigurationRepository) *ResetMaturityScaleHandler {
	return &ResetMaturityScaleHandler{
		repository: repository,
	}
}

func (h *ResetMaturityScaleHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.ResetMaturityScale)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	config, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	modifiedBy, err := valueobjects.NewUserEmail(command.ModifiedBy)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := config.ResetToDefaults(modifiedBy); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, config); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
