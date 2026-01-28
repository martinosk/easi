package services

import (
	"context"
	"testing"

	"easi/backend/internal/capabilitymapping/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRatingLookup struct {
	ratings map[string]*RatingInfo
}

func newMockRatingLookup() *mockRatingLookup {
	return &mockRatingLookup{
		ratings: make(map[string]*RatingInfo),
	}
}

func (m *mockRatingLookup) GetRating(ctx context.Context, capabilityID valueobjects.CapabilityID, pillarID valueobjects.PillarID, businessDomainID valueobjects.BusinessDomainID) (*RatingInfo, error) {
	key := capabilityID.Value() + ":" + pillarID.Value() + ":" + businessDomainID.Value()
	rating, ok := m.ratings[key]
	if !ok {
		return nil, nil
	}
	return rating, nil
}

func (m *mockRatingLookup) addRating(capabilityID valueobjects.CapabilityID, pillarID valueobjects.PillarID, businessDomainID valueobjects.BusinessDomainID, importance int, capabilityName string) {
	key := capabilityID.Value() + ":" + pillarID.Value() + ":" + businessDomainID.Value()
	imp, _ := valueobjects.NewImportance(importance)
	m.ratings[key] = &RatingInfo{
		Importance:     imp,
		CapabilityID:   capabilityID,
		CapabilityName: capabilityName,
		Rationale:      "",
	}
}

func newTestPillarID() valueobjects.PillarID {
	return valueobjects.NewPillarID()
}

func newTestDomainID() valueobjects.BusinessDomainID {
	return valueobjects.NewBusinessDomainID()
}

func TestHierarchicalRatingResolver_Scenario1_ChildCapabilityHasRating_UseChildsRating(t *testing.T) {
	lookup := newMockCapabilityLookup()
	ratingLookup := newMockRatingLookup()
	hierarchyService := NewCapabilityHierarchyService(lookup)
	resolver := NewHierarchicalRatingResolver(hierarchyService, ratingLookup, lookup)

	l1ID := valueobjects.NewCapabilityID()
	l2ID := valueobjects.NewCapabilityID()
	pillarID := newTestPillarID()
	domainID := newTestDomainID()

	lookup.addCapability(l1ID, valueobjects.LevelL1, valueobjects.CapabilityID{})
	lookup.addCapability(l2ID, valueobjects.LevelL2, l1ID)

	ratingLookup.addRating(l1ID, pillarID, domainID, 3, "Payment Processing")
	ratingLookup.addRating(l2ID, pillarID, domainID, 5, "Card Payments")

	result, err := resolver.ResolveEffectiveImportance(context.Background(), l2ID, pillarID, domainID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 5, result.EffectiveImportance.Importance().Value())
	assert.Equal(t, l2ID.Value(), result.EffectiveImportance.SourceCapabilityID().Value())
	assert.False(t, result.EffectiveImportance.IsInherited())
	assert.Equal(t, "Card Payments", result.SourceCapabilityName)
}

func TestHierarchicalRatingResolver_Scenario2_ChildHasNoRating_UseParentsRating(t *testing.T) {
	lookup := newMockCapabilityLookup()
	ratingLookup := newMockRatingLookup()
	hierarchyService := NewCapabilityHierarchyService(lookup)
	resolver := NewHierarchicalRatingResolver(hierarchyService, ratingLookup, lookup)

	l1ID := valueobjects.NewCapabilityID()
	l2ID := valueobjects.NewCapabilityID()
	pillarID := newTestPillarID()
	domainID := newTestDomainID()

	lookup.addCapability(l1ID, valueobjects.LevelL1, valueobjects.CapabilityID{})
	lookup.addCapability(l2ID, valueobjects.LevelL2, l1ID)

	ratingLookup.addRating(l1ID, pillarID, domainID, 4, "Payment Processing")

	result, err := resolver.ResolveEffectiveImportance(context.Background(), l2ID, pillarID, domainID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 4, result.EffectiveImportance.Importance().Value())
	assert.Equal(t, l1ID.Value(), result.EffectiveImportance.SourceCapabilityID().Value())
	assert.True(t, result.EffectiveImportance.IsInherited())
	assert.Equal(t, "Payment Processing", result.SourceCapabilityName)
}

func TestHierarchicalRatingResolver_Scenario3_DeepHierarchyRatingInheritance(t *testing.T) {
	lookup := newMockCapabilityLookup()
	ratingLookup := newMockRatingLookup()
	hierarchyService := NewCapabilityHierarchyService(lookup)
	resolver := NewHierarchicalRatingResolver(hierarchyService, ratingLookup, lookup)

	l1ID := valueobjects.NewCapabilityID()
	l2ID := valueobjects.NewCapabilityID()
	l3ID := valueobjects.NewCapabilityID()
	pillarID := newTestPillarID()
	domainID := newTestDomainID()

	lookup.addCapability(l1ID, valueobjects.LevelL1, valueobjects.CapabilityID{})
	lookup.addCapability(l2ID, valueobjects.LevelL2, l1ID)
	lookup.addCapability(l3ID, valueobjects.LevelL3, l2ID)

	ratingLookup.addRating(l1ID, pillarID, domainID, 5, "Customer Management")

	result, err := resolver.ResolveEffectiveImportance(context.Background(), l3ID, pillarID, domainID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 5, result.EffectiveImportance.Importance().Value())
	assert.Equal(t, l1ID.Value(), result.EffectiveImportance.SourceCapabilityID().Value())
	assert.True(t, result.EffectiveImportance.IsInherited())
	assert.Equal(t, "Customer Management", result.SourceCapabilityName)
}

func TestHierarchicalRatingResolver_Scenario4_MidHierarchyRatingTakesPrecedenceOverParent(t *testing.T) {
	lookup := newMockCapabilityLookup()
	ratingLookup := newMockRatingLookup()
	hierarchyService := NewCapabilityHierarchyService(lookup)
	resolver := NewHierarchicalRatingResolver(hierarchyService, ratingLookup, lookup)

	l1ID := valueobjects.NewCapabilityID()
	l2ID := valueobjects.NewCapabilityID()
	l3ID := valueobjects.NewCapabilityID()
	pillarID := newTestPillarID()
	domainID := newTestDomainID()

	lookup.addCapability(l1ID, valueobjects.LevelL1, valueobjects.CapabilityID{})
	lookup.addCapability(l2ID, valueobjects.LevelL2, l1ID)
	lookup.addCapability(l3ID, valueobjects.LevelL3, l2ID)

	ratingLookup.addRating(l1ID, pillarID, domainID, 3, "Customer Management")
	ratingLookup.addRating(l2ID, pillarID, domainID, 5, "Customer Onboarding")

	result, err := resolver.ResolveEffectiveImportance(context.Background(), l3ID, pillarID, domainID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 5, result.EffectiveImportance.Importance().Value())
	assert.Equal(t, l2ID.Value(), result.EffectiveImportance.SourceCapabilityID().Value())
	assert.True(t, result.EffectiveImportance.IsInherited())
	assert.Equal(t, "Customer Onboarding", result.SourceCapabilityName)
}

func TestHierarchicalRatingResolver_Scenario5_ApplicationRealizesMultipleCapabilitiesInSameChain_ShowAllGaps(t *testing.T) {
	lookup := newMockCapabilityLookup()
	ratingLookup := newMockRatingLookup()
	hierarchyService := NewCapabilityHierarchyService(lookup)
	resolver := NewHierarchicalRatingResolver(hierarchyService, ratingLookup, lookup)

	l1ID := valueobjects.NewCapabilityID()
	l2ID := valueobjects.NewCapabilityID()
	pillarID := newTestPillarID()
	domainID := newTestDomainID()

	lookup.addCapability(l1ID, valueobjects.LevelL1, valueobjects.CapabilityID{})
	lookup.addCapability(l2ID, valueobjects.LevelL2, l1ID)

	ratingLookup.addRating(l1ID, pillarID, domainID, 4, "Payment Processing")
	ratingLookup.addRating(l2ID, pillarID, domainID, 5, "Card Payments")

	resultL1, err := resolver.ResolveEffectiveImportance(context.Background(), l1ID, pillarID, domainID)
	require.NoError(t, err)
	require.NotNil(t, resultL1)
	assert.Equal(t, 4, resultL1.EffectiveImportance.Importance().Value())
	assert.False(t, resultL1.EffectiveImportance.IsInherited())

	resultL2, err := resolver.ResolveEffectiveImportance(context.Background(), l2ID, pillarID, domainID)
	require.NoError(t, err)
	require.NotNil(t, resultL2)
	assert.Equal(t, 5, resultL2.EffectiveImportance.Importance().Value())
	assert.False(t, resultL2.EffectiveImportance.IsInherited())
}

func TestHierarchicalRatingResolver_Scenario6_ApplicationRealizesMultipleCapabilities_MixedRatedAndUnrated(t *testing.T) {
	lookup := newMockCapabilityLookup()
	ratingLookup := newMockRatingLookup()
	hierarchyService := NewCapabilityHierarchyService(lookup)
	resolver := NewHierarchicalRatingResolver(hierarchyService, ratingLookup, lookup)

	l1ID := valueobjects.NewCapabilityID()
	l2ID := valueobjects.NewCapabilityID()
	l3ID := valueobjects.NewCapabilityID()
	pillarID := newTestPillarID()
	domainID := newTestDomainID()

	lookup.addCapability(l1ID, valueobjects.LevelL1, valueobjects.CapabilityID{})
	lookup.addCapability(l2ID, valueobjects.LevelL2, l1ID)
	lookup.addCapability(l3ID, valueobjects.LevelL3, l2ID)

	ratingLookup.addRating(l1ID, pillarID, domainID, 4, "Customer Management")
	ratingLookup.addRating(l2ID, pillarID, domainID, 5, "Customer Onboarding")

	resultL2, err := resolver.ResolveEffectiveImportance(context.Background(), l2ID, pillarID, domainID)
	require.NoError(t, err)
	require.NotNil(t, resultL2)
	assert.Equal(t, 5, resultL2.EffectiveImportance.Importance().Value())
	assert.False(t, resultL2.EffectiveImportance.IsInherited())
	assert.Equal(t, "Customer Onboarding", resultL2.SourceCapabilityName)

	resultL3, err := resolver.ResolveEffectiveImportance(context.Background(), l3ID, pillarID, domainID)
	require.NoError(t, err)
	require.NotNil(t, resultL3)
	assert.Equal(t, 5, resultL3.EffectiveImportance.Importance().Value())
	assert.True(t, resultL3.EffectiveImportance.IsInherited())
	assert.Equal(t, "Customer Onboarding", resultL3.SourceCapabilityName)
}

func TestHierarchicalRatingResolver_Scenario7_CapabilityChainWithNoRatingsAnywhere(t *testing.T) {
	lookup := newMockCapabilityLookup()
	ratingLookup := newMockRatingLookup()
	hierarchyService := NewCapabilityHierarchyService(lookup)
	resolver := NewHierarchicalRatingResolver(hierarchyService, ratingLookup, lookup)

	l1ID := valueobjects.NewCapabilityID()
	l2ID := valueobjects.NewCapabilityID()
	pillarID := newTestPillarID()
	domainID := newTestDomainID()

	lookup.addCapability(l1ID, valueobjects.LevelL1, valueobjects.CapabilityID{})
	lookup.addCapability(l2ID, valueobjects.LevelL2, l1ID)

	result, err := resolver.ResolveEffectiveImportance(context.Background(), l2ID, pillarID, domainID)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestHierarchicalRatingResolver_Scenario8_ApplicationRealizesCapabilitiesInDifferentBranches(t *testing.T) {
	lookup := newMockCapabilityLookup()
	ratingLookup := newMockRatingLookup()
	hierarchyService := NewCapabilityHierarchyService(lookup)
	resolver := NewHierarchicalRatingResolver(hierarchyService, ratingLookup, lookup)

	salesL1ID := valueobjects.NewCapabilityID()
	leadManagementL2ID := valueobjects.NewCapabilityID()
	marketingL1ID := valueobjects.NewCapabilityID()
	campaignManagementL2ID := valueobjects.NewCapabilityID()
	pillarID := newTestPillarID()
	domainID := newTestDomainID()

	lookup.addCapability(salesL1ID, valueobjects.LevelL1, valueobjects.CapabilityID{})
	lookup.addCapability(leadManagementL2ID, valueobjects.LevelL2, salesL1ID)
	lookup.addCapability(marketingL1ID, valueobjects.LevelL1, valueobjects.CapabilityID{})
	lookup.addCapability(campaignManagementL2ID, valueobjects.LevelL2, marketingL1ID)

	ratingLookup.addRating(salesL1ID, pillarID, domainID, 4, "Sales")
	ratingLookup.addRating(leadManagementL2ID, pillarID, domainID, 3, "Lead Management")
	ratingLookup.addRating(marketingL1ID, pillarID, domainID, 5, "Marketing")

	resultLeadManagement, err := resolver.ResolveEffectiveImportance(context.Background(), leadManagementL2ID, pillarID, domainID)
	require.NoError(t, err)
	require.NotNil(t, resultLeadManagement)
	assert.Equal(t, 3, resultLeadManagement.EffectiveImportance.Importance().Value())
	assert.False(t, resultLeadManagement.EffectiveImportance.IsInherited())
	assert.Equal(t, "Lead Management", resultLeadManagement.SourceCapabilityName)

	resultCampaignManagement, err := resolver.ResolveEffectiveImportance(context.Background(), campaignManagementL2ID, pillarID, domainID)
	require.NoError(t, err)
	require.NotNil(t, resultCampaignManagement)
	assert.Equal(t, 5, resultCampaignManagement.EffectiveImportance.Importance().Value())
	assert.True(t, resultCampaignManagement.EffectiveImportance.IsInherited())
	assert.Equal(t, "Marketing", resultCampaignManagement.SourceCapabilityName)
}

func TestHierarchicalRatingResolver_Scenario9_OnlyParentRated_ChildRealized_ShowsUnderChildName(t *testing.T) {
	lookup := newMockCapabilityLookup()
	ratingLookup := newMockRatingLookup()
	hierarchyService := NewCapabilityHierarchyService(lookup)
	resolver := NewHierarchicalRatingResolver(hierarchyService, ratingLookup, lookup)

	financeL1ID := valueobjects.NewCapabilityID()
	accountsPayableL2ID := valueobjects.NewCapabilityID()
	pillarID := newTestPillarID()
	domainID := newTestDomainID()

	lookup.addCapability(financeL1ID, valueobjects.LevelL1, valueobjects.CapabilityID{})
	lookup.addCapability(accountsPayableL2ID, valueobjects.LevelL2, financeL1ID)

	ratingLookup.addRating(financeL1ID, pillarID, domainID, 5, "Finance")

	result, err := resolver.ResolveEffectiveImportance(context.Background(), accountsPayableL2ID, pillarID, domainID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 5, result.EffectiveImportance.Importance().Value())
	assert.Equal(t, financeL1ID.Value(), result.EffectiveImportance.SourceCapabilityID().Value())
	assert.True(t, result.EffectiveImportance.IsInherited())
	assert.Equal(t, "Finance", result.SourceCapabilityName)
}

func TestHierarchicalRatingResolver_Scenario10_PillarSpecificRatingInheritance(t *testing.T) {
	lookup := newMockCapabilityLookup()
	ratingLookup := newMockRatingLookup()
	hierarchyService := NewCapabilityHierarchyService(lookup)
	resolver := NewHierarchicalRatingResolver(hierarchyService, ratingLookup, lookup)

	hrL1ID := valueobjects.NewCapabilityID()
	recruitmentL2ID := valueobjects.NewCapabilityID()
	alwaysOnPillarID := newTestPillarID()
	transformPillarID := newTestPillarID()
	domainID := newTestDomainID()

	lookup.addCapability(hrL1ID, valueobjects.LevelL1, valueobjects.CapabilityID{})
	lookup.addCapability(recruitmentL2ID, valueobjects.LevelL2, hrL1ID)

	ratingLookup.addRating(hrL1ID, alwaysOnPillarID, domainID, 5, "HR")
	ratingLookup.addRating(hrL1ID, transformPillarID, domainID, 2, "HR")
	ratingLookup.addRating(recruitmentL2ID, transformPillarID, domainID, 4, "Recruitment")

	resultAlwaysOn, err := resolver.ResolveEffectiveImportance(context.Background(), recruitmentL2ID, alwaysOnPillarID, domainID)
	require.NoError(t, err)
	require.NotNil(t, resultAlwaysOn)
	assert.Equal(t, 5, resultAlwaysOn.EffectiveImportance.Importance().Value())
	assert.True(t, resultAlwaysOn.EffectiveImportance.IsInherited())
	assert.Equal(t, "HR", resultAlwaysOn.SourceCapabilityName)

	resultTransform, err := resolver.ResolveEffectiveImportance(context.Background(), recruitmentL2ID, transformPillarID, domainID)
	require.NoError(t, err)
	require.NotNil(t, resultTransform)
	assert.Equal(t, 4, resultTransform.EffectiveImportance.Importance().Value())
	assert.False(t, resultTransform.EffectiveImportance.IsInherited())
	assert.Equal(t, "Recruitment", resultTransform.SourceCapabilityName)
}

func TestHierarchicalRatingResolver_CapabilityNotFound_ReturnsError(t *testing.T) {
	lookup := newMockCapabilityLookup()
	ratingLookup := newMockRatingLookup()
	hierarchyService := NewCapabilityHierarchyService(lookup)
	resolver := NewHierarchicalRatingResolver(hierarchyService, ratingLookup, lookup)

	unknownID := valueobjects.NewCapabilityID()
	pillarID := newTestPillarID()
	domainID := newTestDomainID()

	_, err := resolver.ResolveEffectiveImportance(context.Background(), unknownID, pillarID, domainID)
	assert.Error(t, err)
	assert.Equal(t, ErrCapabilityNotFound, err)
}

func TestHierarchicalRatingResolver_DirectRatingOnRootCapability_NotInherited(t *testing.T) {
	lookup := newMockCapabilityLookup()
	ratingLookup := newMockRatingLookup()
	hierarchyService := NewCapabilityHierarchyService(lookup)
	resolver := NewHierarchicalRatingResolver(hierarchyService, ratingLookup, lookup)

	l1ID := valueobjects.NewCapabilityID()
	pillarID := newTestPillarID()
	domainID := newTestDomainID()

	lookup.addCapability(l1ID, valueobjects.LevelL1, valueobjects.CapabilityID{})
	ratingLookup.addRating(l1ID, pillarID, domainID, 3, "Root Capability")

	result, err := resolver.ResolveEffectiveImportance(context.Background(), l1ID, pillarID, domainID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 3, result.EffectiveImportance.Importance().Value())
	assert.Equal(t, l1ID.Value(), result.EffectiveImportance.SourceCapabilityID().Value())
	assert.False(t, result.EffectiveImportance.IsInherited())
}
