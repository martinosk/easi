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

func (h *UnassignCapabilityFromDomainHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.UnassignCapabilityFromDomain)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	assignment, err := h.repository.GetByID(ctx, command.AssignmentID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := assignment.Unassign(); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, assignment); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), nil
}
