package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type UpdateBusinessDomainHandler struct {
	repository *repositories.BusinessDomainRepository
	readModel  *readmodels.BusinessDomainReadModel
}

func NewUpdateBusinessDomainHandler(
	repository *repositories.BusinessDomainRepository,
	readModel *readmodels.BusinessDomainReadModel,
) *UpdateBusinessDomainHandler {
	return &UpdateBusinessDomainHandler{
		repository: repository,
		readModel:  readModel,
	}
}

func (h *UpdateBusinessDomainHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.UpdateBusinessDomain)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	domain, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrBusinessDomainNotFound) {
			return ErrBusinessDomainNotFound
		}
		return err
	}

	exists, err := h.readModel.NameExists(ctx, command.Name, command.ID)
	if err != nil {
		return err
	}
	if exists {
		return ErrBusinessDomainNameExists
	}

	name, err := valueobjects.NewDomainName(command.Name)
	if err != nil {
		return err
	}

	description := valueobjects.NewDescription(command.Description)

	if err := domain.Update(name, description); err != nil {
		return err
	}

	return h.repository.Save(ctx, domain)
}
