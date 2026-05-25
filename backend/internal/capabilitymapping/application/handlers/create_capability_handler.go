package handlers

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

type CreateCapabilityRepository interface {
	Save(ctx context.Context, capability *aggregates.Capability) error
}

type CreateCapabilityHandler struct {
	repository CreateCapabilityRepository
}

func NewCreateCapabilityHandler(repository CreateCapabilityRepository) *CreateCapabilityHandler {
	return &CreateCapabilityHandler{
		repository: repository,
	}
}

func (h *CreateCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CreateCapability)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	name, err := valueobjects.NewCapabilityName(command.Name)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	description, err := valueobjects.NewDescription(command.Description)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	level, err := valueobjects.NewCapabilityLevel(command.Level)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	var parentID valueobjects.CapabilityID
	if command.ParentID != "" {
		parentID, err = valueobjects.NewCapabilityIDFromString(command.ParentID)
		if err != nil {
			return cqrs.EmptyResult(), err
		}
	}

	capability, err := aggregates.NewCapability(name, description, parentID, level)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, capability); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(capability.ID()), nil
}
