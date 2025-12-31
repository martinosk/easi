package handlers

import (
	"context"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"
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

	exists, err := h.readModel.NameExists(ctx, command.Name, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if exists {
		return cqrs.EmptyResult(), ErrEnterpriseCapabilityNameExists
	}

	capability, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	name, err := valueobjects.NewEnterpriseCapabilityName(command.Name)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	description, err := valueobjects.NewDescription(command.Description)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	category, err := valueobjects.NewCategory(command.Category)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := capability.Update(name, description, category); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, capability); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
