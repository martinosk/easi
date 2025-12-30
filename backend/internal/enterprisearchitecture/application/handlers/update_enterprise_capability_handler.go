package handlers

import (
	"context"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"
	"easi/backend/internal/enterprisearchitecture/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type UpdateEnterpriseCapabilityHandler struct {
	repository *repositories.EnterpriseCapabilityRepository
	readModel  *readmodels.EnterpriseCapabilityReadModel
}

func NewUpdateEnterpriseCapabilityHandler(
	repository *repositories.EnterpriseCapabilityRepository,
	readModel *readmodels.EnterpriseCapabilityReadModel,
) *UpdateEnterpriseCapabilityHandler {
	return &UpdateEnterpriseCapabilityHandler{
		repository: repository,
		readModel:  readModel,
	}
}

func (h *UpdateEnterpriseCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.UpdateEnterpriseCapability)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	exists, err := h.readModel.NameExists(ctx, command.Name, command.ID)
	if err != nil {
		return err
	}
	if exists {
		return ErrEnterpriseCapabilityNameExists
	}

	capability, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
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

	if err := capability.Update(name, description, category); err != nil {
		return err
	}

	return h.repository.Save(ctx, capability)
}
