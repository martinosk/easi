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

func (h *ResetMaturityScaleHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.ResetMaturityScale)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	config, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	modifiedBy, err := valueobjects.NewUserEmail(command.ModifiedBy)
	if err != nil {
		return err
	}

	if err := config.ResetToDefaults(modifiedBy); err != nil {
		return err
	}

	return h.repository.Save(ctx, config)
}
