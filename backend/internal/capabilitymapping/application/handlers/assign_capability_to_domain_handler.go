package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

var (
	ErrBusinessDomainNotFound  = errors.New("business domain not found")
	ErrCapabilityNotFound      = errors.New("capability not found")
	ErrAssignmentAlreadyExists = errors.New("this capability is already assigned to this business domain")
)

type AssignCapabilityAssignmentRepository interface {
	Save(ctx context.Context, assignment *aggregates.BusinessDomainAssignment) error
}

type AssignCapabilityCapabilityRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.Capability, error)
}

type AssignCapabilityDomainReader interface {
	GetByID(ctx context.Context, id string) (*readmodels.BusinessDomainDTO, error)
}

type AssignCapabilityAssignmentReader interface {
	AssignmentExists(ctx context.Context, domainID, capabilityID string) (bool, error)
}

type AssignCapabilityToDomainHandler struct {
	assignmentRepo   AssignCapabilityAssignmentRepository
	capabilityRepo   AssignCapabilityCapabilityRepository
	domainReader     AssignCapabilityDomainReader
	assignmentReader AssignCapabilityAssignmentReader
}

func NewAssignCapabilityToDomainHandler(
	assignmentRepo AssignCapabilityAssignmentRepository,
	capabilityRepo AssignCapabilityCapabilityRepository,
	domainReader AssignCapabilityDomainReader,
	assignmentReader AssignCapabilityAssignmentReader,
) *AssignCapabilityToDomainHandler {
	return &AssignCapabilityToDomainHandler{
		assignmentRepo:   assignmentRepo,
		capabilityRepo:   capabilityRepo,
		domainReader:     domainReader,
		assignmentReader: assignmentReader,
	}
}

func (h *AssignCapabilityToDomainHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.AssignCapabilityToDomain)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	businessDomainID, err := valueobjects.NewBusinessDomainIDFromString(command.BusinessDomainID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	capabilityID, err := valueobjects.NewCapabilityIDFromString(command.CapabilityID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.validateBusinessDomainExists(ctx, command.BusinessDomainID); err != nil {
		return cqrs.EmptyResult(), err
	}

	capability, err := h.capabilityRepo.GetByID(ctx, command.CapabilityID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := capability.CanBeAssignedToDomain(); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.validateAssignmentDoesNotExist(ctx, command.BusinessDomainID, command.CapabilityID); err != nil {
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
