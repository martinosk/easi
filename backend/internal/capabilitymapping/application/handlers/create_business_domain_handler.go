package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

var ErrBusinessDomainNameExists = errors.New("business domain with this name already exists")

type CreateBusinessDomainRepository interface {
	Save(ctx context.Context, domain *aggregates.BusinessDomain) error
}

type CreateBusinessDomainReadModel interface {
	NameExists(ctx context.Context, name, excludeID string) (bool, error)
}

type CreateBusinessDomainHandler struct {
	repository CreateBusinessDomainRepository
	readModel  CreateBusinessDomainReadModel
}

func NewCreateBusinessDomainHandler(
	repository CreateBusinessDomainRepository,
	readModel CreateBusinessDomainReadModel,
) *CreateBusinessDomainHandler {
	return &CreateBusinessDomainHandler{
		repository: repository,
		readModel:  readModel,
	}
}

func (h *CreateBusinessDomainHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CreateBusinessDomain)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	exists, err := h.readModel.NameExists(ctx, command.Name, "")
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if exists {
		return cqrs.EmptyResult(), ErrBusinessDomainNameExists
	}

	name, err := valueobjects.NewDomainName(command.Name)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	description, err := valueobjects.NewDescription(command.Description)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	domain, err := aggregates.NewBusinessDomain(name, description)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, domain); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(domain.ID()), nil
}
