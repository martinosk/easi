package services

import (
	"context"
	"testing"

	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCapabilityReparentingService_DetermineNewLevel_Root(t *testing.T) {
	lookup := newMockCapabilityLookup()
	service := NewCapabilityReparentingService(lookup)

	capID := valueobjects.NewCapabilityID()
	lookup.addCapability(capID, valueobjects.LevelL2, valueobjects.NewCapabilityID())

	level, err := service.DetermineNewLevel(context.Background(), capID, valueobjects.CapabilityID{}, valueobjects.LevelL1)
	require.NoError(t, err)
	assert.Equal(t, valueobjects.LevelL1, level)
}

func TestCapabilityReparentingService_DetermineNewLevel_CircularReference(t *testing.T) {
	lookup := newMockCapabilityLookup()
	service := NewCapabilityReparentingService(lookup)

	parentID := valueobjects.NewCapabilityID()
	childID := valueobjects.NewCapabilityID()

	lookup.addCapability(parentID, valueobjects.LevelL1, valueobjects.CapabilityID{})
	lookup.addCapability(childID, valueobjects.LevelL2, parentID)
	lookup.addChild(parentID, childID)

	_, err := service.DetermineNewLevel(context.Background(), parentID, childID, valueobjects.LevelL2)
	assert.Error(t, err)
	assert.Equal(t, aggregates.ErrWouldCreateCircularReference, err)
}

func TestCapabilityReparentingService_DetermineNewLevel_MaxDepthExceeded(t *testing.T) {
	lookup := newMockCapabilityLookup()
	service := NewCapabilityReparentingService(lookup)

	rootID := valueobjects.NewCapabilityID()
	childID := valueobjects.NewCapabilityID()
	grandChildID := valueobjects.NewCapabilityID()

	lookup.addCapability(rootID, valueobjects.LevelL1, valueobjects.CapabilityID{})
	lookup.addCapability(childID, valueobjects.LevelL2, rootID)
	lookup.addCapability(grandChildID, valueobjects.LevelL3, childID)
	lookup.addChild(rootID, childID)
	lookup.addChild(childID, grandChildID)

	parentLevel := valueobjects.LevelL3
	_, err := service.DetermineNewLevel(context.Background(), rootID, valueobjects.NewCapabilityID(), parentLevel)
	assert.Error(t, err)
	assert.Equal(t, aggregates.ErrWouldExceedMaximumDepth, err)
}

func TestCapabilityReparentingService_CalculateChildLevel(t *testing.T) {
	lookup := newMockCapabilityLookup()
	service := NewCapabilityReparentingService(lookup)

	level, err := service.CalculateChildLevel(valueobjects.LevelL2)
	require.NoError(t, err)
	assert.Equal(t, valueobjects.LevelL3, level)
}
