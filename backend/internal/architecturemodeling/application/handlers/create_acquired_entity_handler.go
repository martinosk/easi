package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type CreateAcquiredEntityHandler struct {
	repository *repositories.AcquiredEntityRepository
}

func NewCreateAcquiredEntityHandler(repository *repositories.AcquiredEntityRepository) *CreateAcquiredEntityHandler {
	return &CreateAcquiredEntityHandler{
		repository: repository,
	}
}

func (h *CreateAcquiredEntityHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CreateAcquiredEntity)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	name, err := valueobjects.NewEntityName(command.Name)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	integrationStatus, err := valueobjects.NewIntegrationStatus(command.IntegrationStatus)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	notes, err := valueobjects.NewNotes(command.Notes)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	entity, err := aggregates.NewAcquiredEntity(name, command.AcquisitionDate, integrationStatus, notes)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, entity); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(entity.ID()), nil
}
