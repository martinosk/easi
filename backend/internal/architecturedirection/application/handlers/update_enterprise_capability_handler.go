package handlers

import (
	"context"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/shared/cqrs"
)

type UpdateCapabilityRepository interface {
	Save(ctx context.Context, capability *aggregates.EnterpriseCapability) error
	GetByID(ctx context.Context, id string) (*aggregates.EnterpriseCapability, error)
}

type UpdateCapabilityReadModel interface {
	NameExists(ctx context.Context, name, excludeID string) (bool, error)
}

type UpdateEnterpriseCapabilityHandler struct {
	repository UpdateCapabilityRepository
	readModel  UpdateCapabilityReadModel
}

func NewUpdateEnterpriseCapabilityHandler(
	repository UpdateCapabilityRepository,
	readModel UpdateCapabilityReadModel,
) *UpdateEnterpriseCapabilityHandler {
	return &UpdateEnterpriseCapabilityHandler{
		repository: repository,
		readModel:  readModel,
	}
}

func (h *UpdateEnterpriseCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.UpdateEnterpriseCapability)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	if err := rejectDuplicateCapabilityName(ctx, h.readModel, command.Name, command.ID); err != nil {
		return cqrs.EmptyResult(), err
	}

	capability, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	details, err := newCapabilityDetails(command.Name, command.Description, command.Category)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := capability.Update(details.name, details.description, details.category); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, capability); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
