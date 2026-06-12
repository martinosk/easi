package services

import (
	"context"
	"errors"
	"fmt"

	"easi/backend/internal/architecturedirection/domain/aggregates"
)

var ErrReferencedEntityNotFound = errors.New("a referenced entity does not exist or is not accessible in this tenant")

var ErrActiveDirectionAlreadyExists = errors.New("an active direction already exists on this enterprise capability")

var ErrEnterpriseCapabilityInactive = errors.New("directions can only be captured on active enterprise capabilities")

type ExistenceCheck func(ctx context.Context, id string) (bool, error)

type ReferenceChecker struct {
	EnterpriseCapabilityExists   ExistenceCheck
	EnterpriseCapabilityIsActive ExistenceCheck
	PhysicalCapabilityExists     ExistenceCheck
}

type ActiveDirectionLookup interface {
	HasActiveDirectionForEnterpriseCapability(ctx context.Context, enterpriseCapabilityID string) (bool, error)
}

type SourceConflict struct {
	CapabilityID             string
	CapabilityName           string
	EnterpriseCapabilityID   string
	EnterpriseCapabilityName string
}

type SourceEligibility interface {
	FirstSourceConflict(ctx context.Context, enterpriseCapabilityID string, sourceCapabilityIDs []string) (*SourceConflict, error)
}

type SourceConflictError struct {
	Conflict SourceConflict
}

func (e *SourceConflictError) Error() string {
	return fmt.Sprintf(
		"Capability '%s' is already an explicit source of an active direction on '%s'. A domain capability may be the explicit source of at most one active direction.",
		e.Conflict.CapabilityName, e.Conflict.EnterpriseCapabilityName,
	)
}

type DirectionReferenceService struct {
	references  *ReferenceChecker
	active      ActiveDirectionLookup
	eligibility SourceEligibility
}

func NewDirectionReferenceService(references *ReferenceChecker, active ActiveDirectionLookup, eligibility SourceEligibility) *DirectionReferenceService {
	return &DirectionReferenceService{references: references, active: active, eligibility: eligibility}
}

func (s *DirectionReferenceService) VerifyCanCapture(ctx context.Context, params aggregates.DraftParams) error {
	if err := s.verifyReferences(ctx, params); err != nil {
		return err
	}
	if err := s.ensureNoActiveDirection(ctx, params.EnterpriseCapabilityID.Value()); err != nil {
		return err
	}
	return s.VerifySourcesEligible(ctx, params.EnterpriseCapabilityID.Value(), sourceIDsOf(params))
}

func (s *DirectionReferenceService) VerifySourcesEligible(ctx context.Context, enterpriseCapabilityID string, sourceCapabilityIDs []string) error {
	if s.eligibility == nil || len(sourceCapabilityIDs) == 0 {
		return nil
	}
	conflict, err := s.eligibility.FirstSourceConflict(ctx, enterpriseCapabilityID, sourceCapabilityIDs)
	if err != nil {
		return fmt.Errorf("check source eligibility for enterprise capability %s: %w", enterpriseCapabilityID, err)
	}
	if conflict != nil {
		return &SourceConflictError{Conflict: *conflict}
	}
	return nil
}

func (s *DirectionReferenceService) verifyReferences(ctx context.Context, params aggregates.DraftParams) error {
	if s.references == nil {
		return nil
	}
	ecID := params.EnterpriseCapabilityID.Value()
	if err := requireExists(ctx, s.references.EnterpriseCapabilityExists, ecID, "enterprise capability"); err != nil {
		return err
	}
	if err := requireActiveEC(ctx, s.references.EnterpriseCapabilityIsActive, ecID); err != nil {
		return err
	}
	return verifyAll(ctx, s.references.PhysicalCapabilityExists, sourceIDsOf(params), "source capability")
}

func (s *DirectionReferenceService) ensureNoActiveDirection(ctx context.Context, enterpriseCapabilityID string) error {
	hasActive, err := s.active.HasActiveDirectionForEnterpriseCapability(ctx, enterpriseCapabilityID)
	if err != nil {
		return err
	}
	if hasActive {
		return ErrActiveDirectionAlreadyExists
	}
	return nil
}

func sourceIDsOf(params aggregates.DraftParams) []string {
	sourceIDs := make([]string, len(params.SourceCapabilityIDs))
	for i, ref := range params.SourceCapabilityIDs {
		sourceIDs[i] = ref.Value()
	}
	return sourceIDs
}

func requireActiveEC(ctx context.Context, check ExistenceCheck, ecID string) error {
	if check == nil {
		return nil
	}
	active, err := check(ctx, ecID)
	if err != nil {
		return fmt.Errorf("verify enterprise capability %s is active: %w", ecID, err)
	}
	if !active {
		return ErrEnterpriseCapabilityInactive
	}
	return nil
}

func verifyAll(ctx context.Context, check ExistenceCheck, ids []string, label string) error {
	for _, id := range ids {
		if err := requireExists(ctx, check, id, label); err != nil {
			return err
		}
	}
	return nil
}

func requireExists(ctx context.Context, check ExistenceCheck, id, label string) error {
	exists, err := check(ctx, id)
	if err != nil {
		return fmt.Errorf("verify %s %s: %w", label, id, err)
	}
	if !exists {
		return fmt.Errorf("%w: %s %s", ErrReferencedEntityNotFound, label, id)
	}
	return nil
}
