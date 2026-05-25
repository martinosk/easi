package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

type UpdateAcquiredEntityRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.AcquiredEntity, error)
	Save(ctx context.Context, entity *aggregates.AcquiredEntity) error
}

type UpdateAcquiredEntityHandler struct {
	repository UpdateAcquiredEntityRepository
}

func NewUpdateAcquiredEntityHandler(repository UpdateAcquiredEntityRepository) *UpdateAcquiredEntityHandler {
	return &UpdateAcquiredEntityHandler{
		repository: repository,
	}
}

func (h *UpdateAcquiredEntityHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.UpdateAcquiredEntity)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	entity, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
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

	if err := entity.Update(name, command.AcquisitionDate, integrationStatus, notes); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, entity); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
