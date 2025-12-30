package handlers

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type CreateCapabilityHandler struct {
	repository *repositories.CapabilityRepository
}

func NewCreateCapabilityHandler(repository *repositories.CapabilityRepository) *CreateCapabilityHandler {
	return &CreateCapabilityHandler{
		repository: repository,
	}
}

func (h *CreateCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.CreateCapability)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	name, err := valueobjects.NewCapabilityName(command.Name)
	if err != nil {
		return err
	}

	description, err := valueobjects.NewDescription(command.Description)
	if err != nil {
		return err
	}

	level, err := valueobjects.NewCapabilityLevel(command.Level)
	if err != nil {
		return err
	}

	var parentID valueobjects.CapabilityID
	if command.ParentID != "" {
		parentID, err = valueobjects.NewCapabilityIDFromString(command.ParentID)
		if err != nil {
			return err
		}
	}

	capability, err := aggregates.NewCapability(name, description, parentID, level)
	if err != nil {
		return err
	}

	command.ID = capability.ID()

	return h.repository.Save(ctx, capability)
}
