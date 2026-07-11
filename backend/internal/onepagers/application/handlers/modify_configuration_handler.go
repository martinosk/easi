package handlers

import (
	"context"

	"easi/backend/internal/onepagers/domain/aggregates"
	"easi/backend/internal/onepagers/domain/valueobjects"
	"easi/backend/internal/onepagers/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type configurationCommand interface {
	cqrs.Command
	ConfigurationID() string
	ModifiedByEmail() string
}

type modifyAction[C configurationCommand] func(config *aggregates.OnePagerConfiguration, cmd C, modifiedBy valueobjects.UserEmail) (string, error)

type modifyConfigurationHandler[C configurationCommand] struct {
	repository *repositories.OnePagerConfigurationRepository
	act        modifyAction[C]
}

func newModifyHandler[C configurationCommand](
	repository *repositories.OnePagerConfigurationRepository,
	act modifyAction[C],
) cqrs.CommandHandler {
	return &modifyConfigurationHandler[C]{repository: repository, act: act}
}

func fieldAction[C configurationCommand](
	fieldID func(C) string,
	act func(*aggregates.OnePagerConfiguration, valueobjects.FieldID, valueobjects.UserEmail) error,
) modifyAction[C] {
	return func(config *aggregates.OnePagerConfiguration, cmd C, modifiedBy valueobjects.UserEmail) (string, error) {
		id, err := valueobjects.NewFieldIDFromString(fieldID(cmd))
		if err != nil {
			return "", err
		}
		return "", act(config, id, modifiedBy)
	}
}

func (h *modifyConfigurationHandler[C]) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(C)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	modifiedBy, err := valueobjects.NewUserEmail(command.ModifiedByEmail())
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	config, err := h.repository.GetByID(ctx, command.ConfigurationID())
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	createdID, err := h.act(config, command, modifiedBy)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, config); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(createdID), nil
}
