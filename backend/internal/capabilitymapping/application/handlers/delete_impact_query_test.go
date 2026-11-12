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
	assert.NotNil(t, impact.Realizations)
	assert.Empty(t, impact.Realizations)
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

func TestDeleteImpactQuery_ListsEachDirectRealizationInScopeOnce(t *testing.T) {
	rootID := "550e8400-e29b-41d4-a716-446655440001"
	childID := "550e8400-e29b-41d4-a716-446655440002"
	grandchildID := "550e8400-e29b-41d4-a716-446655440003"

	childCapID, _ := valueobjects.NewCapabilityIDFromString(childID)
	grandchildCapID, _ := valueobjects.NewCapabilityIDFromString(grandchildID)
	hierarchy := &mockCascadeHierarchyService{
		descendants: map[string][]valueobjects.CapabilityID{rootID: {childCapID, grandchildCapID}},
	}

	realizationRM := &mockCascadeRealizationRM{
		realizationsByCapability: map[string][]readmodels.RealizationDTO{
			rootID: {
				{ID: "inherited-root", CapabilityID: rootID, ComponentID: "billing", ComponentName: "Billing", Origin: readmodels.OriginInherited, SourceRealizationID: "billing-direct", LinkedAt: time.Now()},
			},
			childID: {
				{ID: "inherited-child", CapabilityID: childID, ComponentID: "billing", ComponentName: "Billing", Origin: readmodels.OriginInherited, SourceRealizationID: "billing-direct", LinkedAt: time.Now()},
			},
			grandchildID: {
				{ID: "billing-direct", CapabilityID: grandchildID, ComponentID: "billing", ComponentName: "Billing", Origin: readmodels.OriginDirect, LinkedAt: time.Now()},
			},
		},
	}

	query := NewDeleteImpactQuery(hierarchy, realizationRM)
	impact, err := query.Execute(context.Background(), rootID)
	require.NoError(t, err)

	require.Len(t, impact.Realizations, 1)
	assert.Equal(t, "billing-direct", impact.Realizations[0].ID)
	assert.Equal(t, grandchildID, impact.Realizations[0].CapabilityID)
	assert.Equal(t, "Billing", impact.Realizations[0].ComponentName)
}

func TestDeleteImpactQuery_RealizedLeaf_ListsItsRealization(t *testing.T) {
	rootID := "550e8400-e29b-41d4-a716-446655440001"

	hierarchy := &mockCascadeHierarchyService{
		descendants: map[string][]valueobjects.CapabilityID{},
	}
	realizationRM := &mockCascadeRealizationRM{
		realizationsByCapability: map[string][]readmodels.RealizationDTO{
			rootID: {
				{ID: "real-1", CapabilityID: rootID, ComponentID: "comp-1", ComponentName: "App 1", Origin: readmodels.OriginDirect, LinkedAt: time.Now()},
			},
		},
	}

	query := NewDeleteImpactQuery(hierarchy, realizationRM)
	impact, err := query.Execute(context.Background(), rootID)
	require.NoError(t, err)

	require.Len(t, impact.Realizations, 1)
	assert.Equal(t, "real-1", impact.Realizations[0].ID)
}
