package services

import (
	"context"
	"errors"

	"easi/backend/internal/capabilitymapping/domain/valueobjects"
)

var (
	ErrCapabilityHasChildren                  = errors.New("cannot delete capability with children")
	ErrCascadeRequiredForChildCapabilities    = errors.New("cascade deletion required for capability with descendants")
)

type CapabilityChildrenChecker interface {
	HasChildren(ctx context.Context, capabilityID valueobjects.CapabilityID) (bool, error)
}

type CapabilityDeletionService interface {
	CanDelete(ctx context.Context, capabilityID valueobjects.CapabilityID) error
}

type capabilityDeletionService struct {
	childrenChecker CapabilityChildrenChecker
}

func NewCapabilityDeletionService(childrenChecker CapabilityChildrenChecker) CapabilityDeletionService {
	return &capabilityDeletionService{
		childrenChecker: childrenChecker,
	}
}

func (s *capabilityDeletionService) CanDelete(ctx context.Context, capabilityID valueobjects.CapabilityID) error {
	hasChildren, err := s.childrenChecker.HasChildren(ctx, capabilityID)
	if err != nil {
		return err
	}

	if hasChildren {
		return ErrCapabilityHasChildren
	}

	return nil
}
