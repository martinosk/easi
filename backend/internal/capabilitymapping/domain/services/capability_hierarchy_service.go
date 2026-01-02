package services

import (
	"context"
	"errors"

	"easi/backend/internal/capabilitymapping/domain/valueobjects"
)

var (
	ErrCapabilityNotFound           = errors.New("capability not found")
	ErrNoL1AncestorFound            = errors.New("no L1 ancestor found in hierarchy")
	ErrWouldCreateCircularHierarchy = errors.New("operation would create circular hierarchy")
)

type CapabilityInfo struct {
	ID       valueobjects.CapabilityID
	Level    valueobjects.CapabilityLevel
	ParentID valueobjects.CapabilityID
}

type CapabilityLookup interface {
	GetCapabilityInfo(ctx context.Context, id valueobjects.CapabilityID) (*CapabilityInfo, error)
	GetChildren(ctx context.Context, parentID valueobjects.CapabilityID) ([]valueobjects.CapabilityID, error)
}

type CapabilityHierarchyService interface {
	FindL1Ancestor(ctx context.Context, capabilityID valueobjects.CapabilityID) (valueobjects.CapabilityID, error)
	GetDescendants(ctx context.Context, capabilityID valueobjects.CapabilityID) ([]valueobjects.CapabilityID, error)
	ValidateHierarchyChange(ctx context.Context, capabilityID valueobjects.CapabilityID, newParentID valueobjects.CapabilityID) error
}

type capabilityHierarchyService struct {
	lookup CapabilityLookup
}

func NewCapabilityHierarchyService(lookup CapabilityLookup) CapabilityHierarchyService {
	return &capabilityHierarchyService{lookup: lookup}
}

func (s *capabilityHierarchyService) FindL1Ancestor(ctx context.Context, capabilityID valueobjects.CapabilityID) (valueobjects.CapabilityID, error) {
	currentID := capabilityID
	visited := make(map[string]bool)

	for {
		if visited[currentID.Value()] {
			return valueobjects.CapabilityID{}, ErrWouldCreateCircularHierarchy
		}
		visited[currentID.Value()] = true

		info, err := s.lookup.GetCapabilityInfo(ctx, currentID)
		if err != nil {
			return valueobjects.CapabilityID{}, err
		}
		if info == nil {
			return valueobjects.CapabilityID{}, ErrCapabilityNotFound
		}

		if info.Level == valueobjects.LevelL1 {
			return info.ID, nil
		}

		if info.ParentID.Value() == "" {
			return info.ID, nil
		}

		currentID = info.ParentID
	}
}

func (s *capabilityHierarchyService) GetDescendants(ctx context.Context, capabilityID valueobjects.CapabilityID) ([]valueobjects.CapabilityID, error) {
	var descendants []valueobjects.CapabilityID

	children, err := s.lookup.GetChildren(ctx, capabilityID)
	if err != nil {
		return nil, err
	}

	for _, childID := range children {
		descendants = append(descendants, childID)

		childDescendants, err := s.GetDescendants(ctx, childID)
		if err != nil {
			return nil, err
		}
		descendants = append(descendants, childDescendants...)
	}

	return descendants, nil
}

func (s *capabilityHierarchyService) ValidateHierarchyChange(ctx context.Context, capabilityID valueobjects.CapabilityID, newParentID valueobjects.CapabilityID) error {
	if newParentID.Value() == "" {
		return nil
	}

	if newParentID.Value() == capabilityID.Value() {
		return ErrWouldCreateCircularHierarchy
	}

	descendants, err := s.GetDescendants(ctx, capabilityID)
	if err != nil {
		return err
	}

	for _, descendantID := range descendants {
		if descendantID.Value() == newParentID.Value() {
			return ErrWouldCreateCircularHierarchy
		}
	}

	return nil
}
