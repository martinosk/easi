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

var (
	ErrBusinessDomainNotFound          = errors.New("business domain not found")
	ErrCapabilityNotFound              = errors.New("capability not found")
	ErrOnlyL1CapabilitiesCanBeAssigned = errors.New("only L1 capabilities can be assigned to business domains")
	ErrAssignmentAlreadyExists         = errors.New("this capability is already assigned to this business domain")
)

type AssignCapabilityToDomainHandler struct {
	assignmentRepo   *repositories.BusinessDomainAssignmentRepository
	domainReader     *readmodels.BusinessDomainReadModel
	capabilityReader *readmodels.CapabilityReadModel
	assignmentReader *readmodels.DomainCapabilityAssignmentReadModel
}

func NewAssignCapabilityToDomainHandler(
	assignmentRepo *repositories.BusinessDomainAssignmentRepository,
	domainReader *readmodels.BusinessDomainReadModel,
	capabilityReader *readmodels.CapabilityReadModel,
	assignmentReader *readmodels.DomainCapabilityAssignmentReadModel,
) *AssignCapabilityToDomainHandler {
	return &AssignCapabilityToDomainHandler{
		assignmentRepo:   assignmentRepo,
		domainReader:     domainReader,
		capabilityReader: capabilityReader,
		assignmentReader: assignmentReader,
	}
}

func (h *AssignCapabilityToDomainHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.AssignCapabilityToDomain)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	if err := h.validateAssignment(ctx, command); err != nil {
		return cqrs.EmptyResult(), err
	}

	businessDomainID, err := valueobjects.NewBusinessDomainIDFromString(command.BusinessDomainID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	capabilityID, err := valueobjects.NewCapabilityIDFromString(command.CapabilityID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	assignment, err := aggregates.AssignCapabilityToDomain(businessDomainID, capabilityID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.assignmentRepo.Save(ctx, assignment); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(assignment.ID()), nil
}

func (h *AssignCapabilityToDomainHandler) validateAssignment(ctx context.Context, command *commands.AssignCapabilityToDomain) error {
	if err := h.validateBusinessDomainExists(ctx, command.BusinessDomainID); err != nil {
		return err
	}

	if err := h.validateCapabilityIsL1(ctx, command.CapabilityID); err != nil {
		return err
	}

	return h.validateAssignmentDoesNotExist(ctx, command.BusinessDomainID, command.CapabilityID)
}

func (h *AssignCapabilityToDomainHandler) validateBusinessDomainExists(ctx context.Context, domainID string) error {
	domain, err := h.domainReader.GetByID(ctx, domainID)
	if err != nil {
		return err
	}
	if domain == nil {
		return ErrBusinessDomainNotFound
	}
	return nil
}

func (h *AssignCapabilityToDomainHandler) validateCapabilityIsL1(ctx context.Context, capabilityID string) error {
	capability, err := h.capabilityReader.GetByID(ctx, capabilityID)
	if err != nil {
		return err
	}
	if capability == nil {
		return ErrCapabilityNotFound
	}
	if capability.Level != "L1" {
		return ErrOnlyL1CapabilitiesCanBeAssigned
	}
	return nil
}

func (h *AssignCapabilityToDomainHandler) validateAssignmentDoesNotExist(ctx context.Context, domainID, capabilityID string) error {
	exists, err := h.assignmentReader.AssignmentExists(ctx, domainID, capabilityID)
	if err != nil {
		return err
	}
	if exists {
		return ErrAssignmentAlreadyExists
	}
	return nil
}
