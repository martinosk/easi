package services

import (
	"context"
	"errors"
	"fmt"

	"easi/backend/internal/architecturedirection/domain/aggregates"
)

var ErrReferencedEntityNotFound = errors.New("a referenced entity does not exist or is not accessible in this tenant")

var ErrActiveDirectionAlreadyExists = errors.New("an active direction already exists on this enterprise capability")

type ExistenceCheck func(ctx context.Context, id string) (bool, error)

type ReferenceChecker struct {
	EnterpriseCapabilityExists ExistenceCheck
	PhysicalCapabilityExists   ExistenceCheck
	BusinessDomainExists       ExistenceCheck
}

type ActiveDirectionLookup interface {
	HasActiveDirectionForEnterpriseCapability(ctx context.Context, enterpriseCapabilityID string) (bool, error)
}

type DirectionReferenceService struct {
	references *ReferenceChecker
	active     ActiveDirectionLookup
}

func NewDirectionReferenceService(references *ReferenceChecker, active ActiveDirectionLookup) *DirectionReferenceService {
	return &DirectionReferenceService{references: references, active: active}
}

func (s *DirectionReferenceService) VerifyCanCapture(ctx context.Context, params aggregates.DraftParams) error {
	if err := s.verifyReferences(ctx, params); err != nil {
		return err
	}
	return s.ensureNoActiveDirection(ctx, params.EnterpriseCapabilityID.Value())
}

func (s *DirectionReferenceService) verifyReferences(ctx context.Context, params aggregates.DraftParams) error {
	if s.references == nil {
		return nil
	}
	if err := requireExists(ctx, s.references.EnterpriseCapabilityExists, params.EnterpriseCapabilityID.Value(), "enterprise capability"); err != nil {
		return err
	}
	sourceIDs := make([]string, len(params.SourceCapabilityIDs))
	for i, ref := range params.SourceCapabilityIDs {
		sourceIDs[i] = ref.Value()
	}
	if err := verifyAll(ctx, s.references.PhysicalCapabilityExists, sourceIDs, "source capability"); err != nil {
		return err
	}
	domainIDs := make([]string, len(params.Placements))
	for i, p := range params.Placements {
		domainIDs[i] = p.TargetBusinessDomainID()
	}
	return verifyAll(ctx, s.references.BusinessDomainExists, domainIDs, "target business domain")
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
