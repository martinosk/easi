package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/shared/cqrs"
)

var ErrEnterpriseCapabilityNameExists = errors.New("enterprise capability with this name already exists")

type CreateCapabilityRepository interface {
	Save(ctx context.Context, capability *aggregates.EnterpriseCapability) error
}

type CreateCapabilityReadModel interface {
	NameExists(ctx context.Context, name, excludeID string) (bool, error)
}

type CreateEnterpriseCapabilityHandler struct {
	repository CreateCapabilityRepository
	readModel  CreateCapabilityReadModel
}

func NewCreateEnterpriseCapabilityHandler(
	repository CreateCapabilityRepository,
	readModel CreateCapabilityReadModel,
) *CreateEnterpriseCapabilityHandler {
	return &CreateEnterpriseCapabilityHandler{
		repository: repository,
		readModel:  readModel,
	}
}

func (h *CreateEnterpriseCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CreateEnterpriseCapability)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	if err := rejectDuplicateCapabilityName(ctx, h.readModel, command.Name, ""); err != nil {
		return cqrs.EmptyResult(), err
	}

	details, err := newCapabilityDetails(command.Name, command.Description, command.Category)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	capability, err := aggregates.NewEnterpriseCapability(details.name, details.description, details.category)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, capability); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(capability.ID()), nil
}
