package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"
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

	exists, err := h.readModel.NameExists(ctx, command.Name, "")
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if exists {
		return cqrs.EmptyResult(), ErrEnterpriseCapabilityNameExists
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

	capability, err := aggregates.NewEnterpriseCapability(name, description, category)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, capability); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(capability.ID()), nil
}
