package handlers

import (
	"context"

	"easi/backend/internal/metamodel/application/commands"
	"easi/backend/internal/metamodel/domain/aggregates"
	"easi/backend/internal/metamodel/domain/valueobjects"
	"easi/backend/internal/metamodel/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type CreateMetaModelConfigurationHandler struct {
	repository *repositories.MetaModelConfigurationRepository
}

func NewCreateMetaModelConfigurationHandler(repository *repositories.MetaModelConfigurationRepository) *CreateMetaModelConfigurationHandler {
	return &CreateMetaModelConfigurationHandler{
		repository: repository,
	}
}

func (h *CreateMetaModelConfigurationHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CreateMetaModelConfiguration)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	tenantID, err := sharedvo.NewTenantID(command.TenantID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	createdBy, err := valueobjects.NewUserEmail(command.CreatedBy)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	config, err := aggregates.NewMetaModelConfiguration(tenantID, createdBy)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, config); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(config.ID()), nil
}
