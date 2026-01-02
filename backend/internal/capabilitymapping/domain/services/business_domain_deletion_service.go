package services

import (
	"context"
	"errors"

	"easi/backend/internal/capabilitymapping/domain/valueobjects"
)

var (
	ErrBusinessDomainHasAssignments = errors.New("cannot delete business domain with active capability assignments")
)

type BusinessDomainAssignmentChecker interface {
	HasAssignments(ctx context.Context, domainID valueobjects.BusinessDomainID) (bool, error)
}

type BusinessDomainDeletionService interface {
	CanDelete(ctx context.Context, domainID valueobjects.BusinessDomainID) error
}

type businessDomainDeletionService struct {
	assignmentChecker BusinessDomainAssignmentChecker
}

func NewBusinessDomainDeletionService(assignmentChecker BusinessDomainAssignmentChecker) BusinessDomainDeletionService {
	return &businessDomainDeletionService{
		assignmentChecker: assignmentChecker,
	}
}

func (s *businessDomainDeletionService) CanDelete(ctx context.Context, domainID valueobjects.BusinessDomainID) error {
	hasAssignments, err := s.assignmentChecker.HasAssignments(ctx, domainID)
	if err != nil {
		return err
	}

	if hasAssignments {
		return ErrBusinessDomainHasAssignments
	}

	return nil
}
