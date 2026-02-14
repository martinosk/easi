package services

import (
	"context"
	"errors"

	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
)

type CapabilityReparentingService interface {
	DetermineNewLevel(ctx context.Context, capabilityID valueobjects.CapabilityID, newParentID valueobjects.CapabilityID, parentLevel valueobjects.CapabilityLevel) (valueobjects.CapabilityLevel, error)
	CalculateChildLevel(parentLevel valueobjects.CapabilityLevel) (valueobjects.CapabilityLevel, error)
}

type capabilityReparentingService struct {
	lookup           CapabilityLookup
	hierarchyService CapabilityHierarchyService
}

func NewCapabilityReparentingService(lookup CapabilityLookup) CapabilityReparentingService {
	return &capabilityReparentingService{
		lookup:           lookup,
		hierarchyService: NewCapabilityHierarchyService(lookup),
	}
}

func (s *capabilityReparentingService) DetermineNewLevel(ctx context.Context, capabilityID valueobjects.CapabilityID, newParentID valueobjects.CapabilityID, parentLevel valueobjects.CapabilityLevel) (valueobjects.CapabilityLevel, error) {
	if newParentID.Value() == "" {
		if err := s.validateDepthConstraints(ctx, capabilityID, valueobjects.LevelL1); err != nil {
			return "", err
		}
		return valueobjects.LevelL1, nil
	}

	if err := s.validateHierarchyChange(ctx, capabilityID, newParentID); err != nil {
		return "", err
	}

	newLevel, err := s.CalculateChildLevel(parentLevel)
	if err != nil {
		return "", err
	}

	if err := s.validateDepthConstraints(ctx, capabilityID, newLevel); err != nil {
		return "", err
	}

	return newLevel, nil
}

func (s *capabilityReparentingService) CalculateChildLevel(parentLevel valueobjects.CapabilityLevel) (valueobjects.CapabilityLevel, error) {
	switch parentLevel {
	case valueobjects.LevelL1:
		return valueobjects.LevelL2, nil
	case valueobjects.LevelL2:
		return valueobjects.LevelL3, nil
	case valueobjects.LevelL3:
		return valueobjects.LevelL4, nil
	default:
		return "", aggregates.ErrWouldExceedMaximumDepth
	}
}

func (s *capabilityReparentingService) validateHierarchyChange(ctx context.Context, capabilityID valueobjects.CapabilityID, newParentID valueobjects.CapabilityID) error {
	if err := s.hierarchyService.ValidateHierarchyChange(ctx, capabilityID, newParentID); err != nil {
		if errors.Is(err, ErrWouldCreateCircularHierarchy) {
			return aggregates.ErrWouldCreateCircularReference
		}
		return err
	}
	return nil
}

func (s *capabilityReparentingService) validateDepthConstraints(ctx context.Context, capabilityID valueobjects.CapabilityID, newLevel valueobjects.CapabilityLevel) error {
	subtreeDepth, err := s.calculateSubtreeDepth(ctx, capabilityID)
	if err != nil {
		return err
	}

	if newLevel.NumericValue()+subtreeDepth > 4 {
		return aggregates.ErrWouldExceedMaximumDepth
	}

	return nil
}

func (s *capabilityReparentingService) calculateSubtreeDepth(ctx context.Context, capabilityID valueobjects.CapabilityID) (int, error) {
	children, err := s.lookup.GetChildren(ctx, capabilityID)
	if err != nil {
		return 0, err
	}

	if len(children) == 0 {
		return 0, nil
	}

	maxChildDepth := 0
	for _, childID := range children {
		childDepth, err := s.calculateSubtreeDepth(ctx, childID)
		if err != nil {
			return 0, err
		}
		if childDepth > maxChildDepth {
			maxChildDepth = childDepth
		}
	}

	return 1 + maxChildDepth, nil
}
