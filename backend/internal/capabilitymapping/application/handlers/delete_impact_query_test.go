package handlers

import (
	"context"
	"testing"
	"time"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteImpactQuery_NoDescendants_ReturnsEmptyImpact(t *testing.T) {
	rootID := "550e8400-e29b-41d4-a716-446655440001"
	hierarchy := &mockCascadeHierarchyService{
		descendants: map[string][]valueobjects.CapabilityID{},
	}
	realizationRM := &mockCascadeRealizationRM{
		realizationsByCapability: map[string][]readmodels.RealizationDTO{},
	}

	query := NewDeleteImpactQuery(hierarchy, realizationRM)
	impact, err := query.Execute(context.Background(), rootID)
	require.NoError(t, err)

	assert.False(t, impact.HasDescendants)
	assert.Empty(t, impact.AffectedCapabilities)
	assert.Empty(t, impact.RealizationsOnDeletedCapabilities)
	assert.Empty(t, impact.RealizationsOnRetainedCapabilities)
}

func TestDeleteImpactQuery_WithDescendants_ReturnsAffectedCapabilities(t *testing.T) {
	rootID := "550e8400-e29b-41d4-a716-446655440001"
	childID := "550e8400-e29b-41d4-a716-446655440002"

	childCapID, _ := valueobjects.NewCapabilityIDFromString(childID)
	hierarchy := &mockCascadeHierarchyService{
		descendants: map[string][]valueobjects.CapabilityID{
			rootID: {childCapID},
		},
	}
	realizationRM := &mockCascadeRealizationRM{
		realizationsByCapability: map[string][]readmodels.RealizationDTO{},
	}

	query := NewDeleteImpactQuery(hierarchy, realizationRM)
	impact, err := query.Execute(context.Background(), rootID)
	require.NoError(t, err)

	assert.True(t, impact.HasDescendants)
	require.Len(t, impact.AffectedCapabilities, 1)
	assert.Equal(t, childID, impact.AffectedCapabilities[0])
}

func TestDeleteImpactQuery_RealizationExclusivelyInScope_GoesToDeletedList(t *testing.T) {
	rootID := "550e8400-e29b-41d4-a716-446655440001"

	hierarchy := &mockCascadeHierarchyService{
		descendants: map[string][]valueobjects.CapabilityID{},
	}

	realizationRM := &mockCascadeRealizationRM{
		realizationsByCapability: map[string][]readmodels.RealizationDTO{
			rootID: {
				{ID: "real-1", CapabilityID: rootID, ComponentID: "comp-1", ComponentName: "App 1", Origin: "Direct", LinkedAt: time.Now()},
			},
		},
		realizationsByComponent: map[string][]readmodels.RealizationDTO{
			"comp-1": {
				{ID: "real-1", CapabilityID: rootID, ComponentID: "comp-1", ComponentName: "App 1", Origin: "Direct", LinkedAt: time.Now()},
			},
		},
	}

	query := NewDeleteImpactQuery(hierarchy, realizationRM)
	impact, err := query.Execute(context.Background(), rootID)
	require.NoError(t, err)

	require.Len(t, impact.RealizationsOnDeletedCapabilities, 1)
	assert.Equal(t, "real-1", impact.RealizationsOnDeletedCapabilities[0].ID)
	assert.Empty(t, impact.RealizationsOnRetainedCapabilities)
}

func TestDeleteImpactQuery_RealizationAlsoOutsideScope_GoesToRetainedList(t *testing.T) {
	rootID := "550e8400-e29b-41d4-a716-446655440001"
	outsideCapID := "550e8400-e29b-41d4-a716-999999999999"

	hierarchy := &mockCascadeHierarchyService{
		descendants: map[string][]valueobjects.CapabilityID{},
	}

	realizationRM := &mockCascadeRealizationRM{
		realizationsByCapability: map[string][]readmodels.RealizationDTO{
			rootID: {
				{ID: "real-1", CapabilityID: rootID, ComponentID: "comp-1", ComponentName: "App 1", Origin: "Direct", LinkedAt: time.Now()},
			},
		},
		realizationsByComponent: map[string][]readmodels.RealizationDTO{
			"comp-1": {
				{ID: "real-1", CapabilityID: rootID, ComponentID: "comp-1", ComponentName: "App 1", Origin: "Direct", LinkedAt: time.Now()},
				{ID: "real-2", CapabilityID: outsideCapID, ComponentID: "comp-1", ComponentName: "App 1", Origin: "Direct", LinkedAt: time.Now()},
			},
		},
	}

	query := NewDeleteImpactQuery(hierarchy, realizationRM)
	impact, err := query.Execute(context.Background(), rootID)
	require.NoError(t, err)

	assert.Empty(t, impact.RealizationsOnDeletedCapabilities)
	require.Len(t, impact.RealizationsOnRetainedCapabilities, 1)
	assert.Equal(t, "real-1", impact.RealizationsOnRetainedCapabilities[0].ID)
}
