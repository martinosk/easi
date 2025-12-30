package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"
	"easi/backend/internal/enterprisearchitecture/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

var ErrEnterpriseCapabilityNameExists = errors.New("enterprise capability with this name already exists")

type CreateEnterpriseCapabilityHandler struct {
	repository *repositories.EnterpriseCapabilityRepository
	readModel  *readmodels.EnterpriseCapabilityReadModel
}

func NewCreateEnterpriseCapabilityHandler(
	repository *repositories.EnterpriseCapabilityRepository,
	readModel *readmodels.EnterpriseCapabilityReadModel,
) *CreateEnterpriseCapabilityHandler {
	return &CreateEnterpriseCapabilityHandler{
		repository: repository,
		readModel:  readModel,
	}
}

func (h *CreateEnterpriseCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.CreateEnterpriseCapability)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	exists, err := h.readModel.NameExists(ctx, command.Name, "")
	if err != nil {
		return err
	}
	if exists {
		return ErrEnterpriseCapabilityNameExists
	}

	name, err := valueobjects.NewEnterpriseCapabilityName(command.Name)
	if err != nil {
		return err
	}

	description, err := valueobjects.NewDescription(command.Description)
	if err != nil {
		return err
	}

	category, err := valueobjects.NewCategory(command.Category)
	if err != nil {
		return err
	}

	capability, err := aggregates.NewEnterpriseCapability(name, description, category)
	if err != nil {
		return err
	}

	command.ID = capability.ID()

	return h.repository.Save(ctx, capability)
}
