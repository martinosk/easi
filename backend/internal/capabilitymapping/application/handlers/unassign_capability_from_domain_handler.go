package handlers

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type UnassignCapabilityFromDomainHandler struct {
	repository *repositories.BusinessDomainAssignmentRepository
}

func NewUnassignCapabilityFromDomainHandler(
	repository *repositories.BusinessDomainAssignmentRepository,
) *UnassignCapabilityFromDomainHandler {
	return &UnassignCapabilityFromDomainHandler{
		repository: repository,
	}
}

func (h *UnassignCapabilityFromDomainHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.UnassignCapabilityFromDomain)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	assignment, err := h.repository.GetByID(ctx, command.AssignmentID)
	if err != nil {
		return err
	}

	if err := assignment.Unassign(); err != nil {
		return err
	}

	return h.repository.Save(ctx, assignment)
}
