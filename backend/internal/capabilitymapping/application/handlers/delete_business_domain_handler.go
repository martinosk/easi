package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

var ErrBusinessDomainHasAssignments = errors.New("cannot delete business domain with assigned capabilities. Unassign capabilities first")

type DeleteBusinessDomainHandler struct {
	repository       *repositories.BusinessDomainRepository
	assignmentReader *readmodels.DomainCapabilityAssignmentReadModel
}

func NewDeleteBusinessDomainHandler(
	repository *repositories.BusinessDomainRepository,
	assignmentReader *readmodels.DomainCapabilityAssignmentReadModel,
) *DeleteBusinessDomainHandler {
	return &DeleteBusinessDomainHandler{
		repository:       repository,
		assignmentReader: assignmentReader,
	}
}

func (h *DeleteBusinessDomainHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.DeleteBusinessDomain)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	assignments, err := h.assignmentReader.GetByDomainID(ctx, command.ID)
	if err != nil {
		return err
	}

	if len(assignments) > 0 {
		return ErrBusinessDomainHasAssignments
	}

	domain, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrBusinessDomainNotFound) {
			return ErrBusinessDomainNotFound
		}
		return err
	}

	if err := domain.Delete(); err != nil {
		return err
	}

	return h.repository.Save(ctx, domain)
}
