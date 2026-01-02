package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/services"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type DeleteBusinessDomainRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.BusinessDomain, error)
	Save(ctx context.Context, domain *aggregates.BusinessDomain) error
}

type DeleteBusinessDomainHandler struct {
	repository      DeleteBusinessDomainRepository
	deletionService services.BusinessDomainDeletionService
}

func NewDeleteBusinessDomainHandler(
	repository DeleteBusinessDomainRepository,
	deletionService services.BusinessDomainDeletionService,
) *DeleteBusinessDomainHandler {
	return &DeleteBusinessDomainHandler{
		repository:      repository,
		deletionService: deletionService,
	}
}

func (h *DeleteBusinessDomainHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.DeleteBusinessDomain)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	domainID, err := valueobjects.NewBusinessDomainIDFromString(command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.deletionService.CanDelete(ctx, domainID); err != nil {
		return cqrs.EmptyResult(), err
	}

	domain, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrBusinessDomainNotFound) {
			return cqrs.EmptyResult(), ErrBusinessDomainNotFound
		}
		return cqrs.EmptyResult(), err
	}

	if err := domain.Delete(); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, domain); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
