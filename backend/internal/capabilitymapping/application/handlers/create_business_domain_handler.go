package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

var ErrBusinessDomainNameExists = errors.New("business domain with this name already exists")

type CreateBusinessDomainHandler struct {
	repository *repositories.BusinessDomainRepository
	readModel  *readmodels.BusinessDomainReadModel
}

func NewCreateBusinessDomainHandler(
	repository *repositories.BusinessDomainRepository,
	readModel *readmodels.BusinessDomainReadModel,
) *CreateBusinessDomainHandler {
	return &CreateBusinessDomainHandler{
		repository: repository,
		readModel:  readModel,
	}
}

func (h *CreateBusinessDomainHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.CreateBusinessDomain)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	exists, err := h.readModel.NameExists(ctx, command.Name, "")
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

	description, err := valueobjects.NewDescription(command.Description)
	if err != nil {
		return err
	}

	domain, err := aggregates.NewBusinessDomain(name, description)
	if err != nil {
		return err
	}

	command.ID = domain.ID()

	return h.repository.Save(ctx, domain)
}
