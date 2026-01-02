package services

import (
	"context"
	"errors"

	"easi/backend/internal/capabilitymapping/domain/valueobjects"
)

var (
	ErrAssignmentNotFound  = errors.New("assignment not found")
	ErrDomainNotFound      = errors.New("business domain not found")
	ErrOnlyL1CanBeAssigned = errors.New("only L1 capabilities can be assigned to business domains")
)

type AssignmentInfo struct {
	AssignmentID     string
	BusinessDomainID valueobjects.BusinessDomainID
	CapabilityID     valueobjects.CapabilityID
}

type AssignmentLookup interface {
	GetByCapabilityID(ctx context.Context, capabilityID valueobjects.CapabilityID) ([]AssignmentInfo, error)
	AssignmentExists(ctx context.Context, domainID valueobjects.BusinessDomainID, capabilityID valueobjects.CapabilityID) (bool, error)
}

type AssignmentCommandExecutor interface {
	Unassign(ctx context.Context, assignmentID string) error
	Assign(ctx context.Context, domainID valueobjects.BusinessDomainID, capabilityID valueobjects.CapabilityID) error
}

type BusinessDomainAssignmentService interface {
	ReassignToL1Ancestor(
		ctx context.Context,
		capabilityID valueobjects.CapabilityID,
		newL1ID valueobjects.CapabilityID,
	) error
	UnassignAllForCapability(ctx context.Context, capabilityID valueobjects.CapabilityID) error
}

type businessDomainAssignmentService struct {
	assignmentLookup  AssignmentLookup
	commandExecutor   AssignmentCommandExecutor
	hierarchyService  CapabilityHierarchyService
}

func NewBusinessDomainAssignmentService(
	assignmentLookup AssignmentLookup,
	commandExecutor AssignmentCommandExecutor,
	hierarchyService CapabilityHierarchyService,
) BusinessDomainAssignmentService {
	return &businessDomainAssignmentService{
		assignmentLookup:  assignmentLookup,
		commandExecutor:   commandExecutor,
		hierarchyService:  hierarchyService,
	}
}

func (s *businessDomainAssignmentService) ReassignToL1Ancestor(
	ctx context.Context,
	capabilityID valueobjects.CapabilityID,
	newL1ID valueobjects.CapabilityID,
) error {
	assignments, err := s.assignmentLookup.GetByCapabilityID(ctx, capabilityID)
	if err != nil {
		return err
	}

	if len(assignments) == 0 {
		return nil
	}

	for _, assignment := range assignments {
		if err := s.commandExecutor.Unassign(ctx, assignment.AssignmentID); err != nil {
			continue
		}

		alreadyAssigned, err := s.assignmentLookup.AssignmentExists(ctx, assignment.BusinessDomainID, newL1ID)
		if err != nil {
			continue
		}

		if alreadyAssigned {
			continue
		}

		if err := s.commandExecutor.Assign(ctx, assignment.BusinessDomainID, newL1ID); err != nil {
			continue
		}
	}

	return nil
}

func (s *businessDomainAssignmentService) UnassignAllForCapability(ctx context.Context, capabilityID valueobjects.CapabilityID) error {
	assignments, err := s.assignmentLookup.GetByCapabilityID(ctx, capabilityID)
	if err != nil {
		return err
	}

	for _, assignment := range assignments {
		if err := s.commandExecutor.Unassign(ctx, assignment.AssignmentID); err != nil {
			continue
		}
	}

	return nil
}
