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

type testRating struct {
	CapabilityID   valueobjects.CapabilityID
	PillarID       valueobjects.PillarID
	DomainID       valueobjects.BusinessDomainID
	Importance     int
	CapabilityName string
}

func (m *mockRatingLookup) addRating(r testRating) {
	key := r.CapabilityID.Value() + ":" + r.PillarID.Value() + ":" + r.DomainID.Value()
	imp, _ := valueobjects.NewImportance(r.Importance)
	m.ratings[key] = &RatingInfo{
		Importance:     imp,
		CapabilityID:   r.CapabilityID,
		CapabilityName: r.CapabilityName,
		Rationale:      "",
	}
}

type resolverTestFixture struct {
	lookup       *mockCapabilityLookup
	ratingLookup *mockRatingLookup
	resolver     HierarchicalRatingResolver
}

func newResolverTestFixture() *resolverTestFixture {
	lookup := newMockCapabilityLookup()
	ratingLookup := newMockRatingLookup()
	hierarchyService := NewCapabilityHierarchyService(lookup)
	resolver := NewHierarchicalRatingResolver(hierarchyService, ratingLookup, lookup)
	return &resolverTestFixture{lookup: lookup, ratingLookup: ratingLookup, resolver: resolver}
}

type capDef struct {
	id       valueobjects.CapabilityID
	level    valueobjects.CapabilityLevel
	parentID valueobjects.CapabilityID
}

func (f *resolverTestFixture) addHierarchy(caps []capDef) {
	for _, c := range caps {
		f.lookup.addCapability(c.id, c.level, c.parentID)
	}
}

func (f *resolverTestFixture) addRatings(ratings []testRating) {
	for _, r := range ratings {
		f.ratingLookup.addRating(r)
	}
}

func (f *resolverTestFixture) resolve(t *testing.T, capID valueobjects.CapabilityID, pillarID valueobjects.PillarID, domainID valueobjects.BusinessDomainID) *ResolvedRating {
	t.Helper()
	result, err := f.resolver.ResolveEffectiveImportance(context.Background(), capID, pillarID, domainID)
	require.NoError(t, err)
	require.NotNil(t, result)
	return result
}

type expectedRating struct {
	importance  int
	sourceCapID valueobjects.CapabilityID
	isInherited bool
	sourceName  string
}

func assertRating(t *testing.T, result *ResolvedRating, expected expectedRating) {
	t.Helper()
	assert.Equal(t, expected.importance, result.EffectiveImportance.Importance().Value())
	assert.Equal(t, expected.sourceCapID.Value(), result.EffectiveImportance.SourceCapabilityID().Value())
	assert.Equal(t, expected.isInherited, result.EffectiveImportance.IsInherited())
	assert.Equal(t, expected.sourceName, result.SourceCapabilityName)
}

func TestHierarchicalRatingResolver_DirectRatings(t *testing.T) {
	l1 := valueobjects.NewCapabilityID()
	l2 := valueobjects.NewCapabilityID()
	pillar := valueobjects.NewPillarID()
	domain := valueobjects.NewBusinessDomainID()
	noParent := valueobjects.CapabilityID{}

	tests := []struct {
		name        string
		caps        []capDef
		ratings     []testRating
		queryCapID  valueobjects.CapabilityID
		importance  int
		sourceCapID valueobjects.CapabilityID
		sourceName  string
	}{
		{
			name:        "child uses own rating over parent",
			caps:        []capDef{{l1, valueobjects.LevelL1, noParent}, {l2, valueobjects.LevelL2, l1}},
			ratings:     []testRating{{l1, pillar, domain, 3, "Payment Processing"}, {l2, pillar, domain, 5, "Card Payments"}},
			queryCapID:  l2,
			importance:  5,
			sourceCapID: l2,
			sourceName:  "Card Payments",
		},
		{
			name:        "root capability uses own rating",
			caps:        []capDef{{l1, valueobjects.LevelL1, noParent}},
			ratings:     []testRating{{l1, pillar, domain, 3, "Root Capability"}},
			queryCapID:  l1,
			importance:  3,
			sourceCapID: l1,
			sourceName:  "Root Capability",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newResolverTestFixture()
			f.addHierarchy(tt.caps)
			f.addRatings(tt.ratings)
			result := f.resolve(t, tt.queryCapID, pillar, domain)
			assertRating(t, result, expectedRating{tt.importance, tt.sourceCapID, false, tt.sourceName})
		})
	}
}

func TestHierarchicalRatingResolver_InheritedRatings(t *testing.T) {
	l1 := valueobjects.NewCapabilityID()
	l2 := valueobjects.NewCapabilityID()
	l3 := valueobjects.NewCapabilityID()
	pillar := valueobjects.NewPillarID()
	domain := valueobjects.NewBusinessDomainID()
	noParent := valueobjects.CapabilityID{}

	tests := []struct {
		name        string
		caps        []capDef
		ratings     []testRating
		queryCapID  valueobjects.CapabilityID
		importance  int
		sourceCapID valueobjects.CapabilityID
		sourceName  string
	}{
		{
			name:        "child inherits from parent",
			caps:        []capDef{{l1, valueobjects.LevelL1, noParent}, {l2, valueobjects.LevelL2, l1}},
			ratings:     []testRating{{l1, pillar, domain, 4, "Payment Processing"}},
			queryCapID:  l2,
			importance:  4,
			sourceCapID: l1,
			sourceName:  "Payment Processing",
		},
		{
			name:        "L3 inherits from L1 ancestor",
			caps:        []capDef{{l1, valueobjects.LevelL1, noParent}, {l2, valueobjects.LevelL2, l1}, {l3, valueobjects.LevelL3, l2}},
			ratings:     []testRating{{l1, pillar, domain, 5, "Customer Management"}},
			queryCapID:  l3,
			importance:  5,
			sourceCapID: l1,
			sourceName:  "Customer Management",
		},
		{
			name:        "mid hierarchy takes precedence over root",
			caps:        []capDef{{l1, valueobjects.LevelL1, noParent}, {l2, valueobjects.LevelL2, l1}, {l3, valueobjects.LevelL3, l2}},
			ratings:     []testRating{{l1, pillar, domain, 3, "Customer Management"}, {l2, pillar, domain, 5, "Customer Onboarding"}},
			queryCapID:  l3,
			importance:  5,
			sourceCapID: l2,
			sourceName:  "Customer Onboarding",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newResolverTestFixture()
			f.addHierarchy(tt.caps)
			f.addRatings(tt.ratings)
			result := f.resolve(t, tt.queryCapID, pillar, domain)
			assertRating(t, result, expectedRating{tt.importance, tt.sourceCapID, true, tt.sourceName})
		})
	}
}

func TestHierarchicalRatingResolver_EdgeCases(t *testing.T) {
	t.Run("no ratings anywhere returns nil", func(t *testing.T) {
		f := newResolverTestFixture()
		l1 := valueobjects.NewCapabilityID()
		l2 := valueobjects.NewCapabilityID()
		pillar := valueobjects.NewPillarID()
		domain := valueobjects.NewBusinessDomainID()

		f.addHierarchy([]capDef{
			{l1, valueobjects.LevelL1, valueobjects.CapabilityID{}},
			{l2, valueobjects.LevelL2, l1},
		})

		result, err := f.resolver.ResolveEffectiveImportance(context.Background(), l2, pillar, domain)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("unknown capability returns error", func(t *testing.T) {
		f := newResolverTestFixture()
		_, err := f.resolver.ResolveEffectiveImportance(
			context.Background(), valueobjects.NewCapabilityID(), valueobjects.NewPillarID(), valueobjects.NewBusinessDomainID(),
		)
		assert.Equal(t, ErrCapabilityNotFound, err)
	})
}

func TestHierarchicalRatingResolver_MultipleResolves(t *testing.T) {
	t.Run("both capabilities rated independently", func(t *testing.T) {
		f := newResolverTestFixture()
		l1 := valueobjects.NewCapabilityID()
		l2 := valueobjects.NewCapabilityID()
		pillar := valueobjects.NewPillarID()
		domain := valueobjects.NewBusinessDomainID()

		f.addHierarchy([]capDef{
			{l1, valueobjects.LevelL1, valueobjects.CapabilityID{}},
			{l2, valueobjects.LevelL2, l1},
		})
		f.addRatings([]testRating{
			{l1, pillar, domain, 4, "Payment Processing"},
			{l2, pillar, domain, 5, "Card Payments"},
		})

		assertRating(t, f.resolve(t, l1, pillar, domain), expectedRating{4, l1, false, "Payment Processing"})
		assertRating(t, f.resolve(t, l2, pillar, domain), expectedRating{5, l2, false, "Card Payments"})
	})

	t.Run("rated L2 uses own and unrated L3 inherits", func(t *testing.T) {
		f := newResolverTestFixture()
		l1 := valueobjects.NewCapabilityID()
		l2 := valueobjects.NewCapabilityID()
		l3 := valueobjects.NewCapabilityID()
		pillar := valueobjects.NewPillarID()
		domain := valueobjects.NewBusinessDomainID()

		f.addHierarchy([]capDef{
			{l1, valueobjects.LevelL1, valueobjects.CapabilityID{}},
			{l2, valueobjects.LevelL2, l1},
			{l3, valueobjects.LevelL3, l2},
		})
		f.addRatings([]testRating{
			{l1, pillar, domain, 4, "Customer Management"},
			{l2, pillar, domain, 5, "Customer Onboarding"},
		})

		assertRating(t, f.resolve(t, l2, pillar, domain), expectedRating{5, l2, false, "Customer Onboarding"})
		assertRating(t, f.resolve(t, l3, pillar, domain), expectedRating{5, l2, true, "Customer Onboarding"})
	})
}

func TestHierarchicalRatingResolver_DifferentBranches(t *testing.T) {
	f := newResolverTestFixture()
	salesL1 := valueobjects.NewCapabilityID()
	leadL2 := valueobjects.NewCapabilityID()
	marketingL1 := valueobjects.NewCapabilityID()
	campaignL2 := valueobjects.NewCapabilityID()
	pillar := valueobjects.NewPillarID()
	domain := valueobjects.NewBusinessDomainID()

	f.addHierarchy([]capDef{
		{salesL1, valueobjects.LevelL1, valueobjects.CapabilityID{}},
		{leadL2, valueobjects.LevelL2, salesL1},
		{marketingL1, valueobjects.LevelL1, valueobjects.CapabilityID{}},
		{campaignL2, valueobjects.LevelL2, marketingL1},
	})
	f.addRatings([]testRating{
		{salesL1, pillar, domain, 4, "Sales"},
		{leadL2, pillar, domain, 3, "Lead Management"},
		{marketingL1, pillar, domain, 5, "Marketing"},
	})

	assertRating(t, f.resolve(t, leadL2, pillar, domain), expectedRating{3, leadL2, false, "Lead Management"})
	assertRating(t, f.resolve(t, campaignL2, pillar, domain), expectedRating{5, marketingL1, true, "Marketing"})
}

func TestHierarchicalRatingResolver_PillarSpecificInheritance(t *testing.T) {
	f := newResolverTestFixture()
	hr := valueobjects.NewCapabilityID()
	recruitment := valueobjects.NewCapabilityID()
	alwaysOn := valueobjects.NewPillarID()
	transform := valueobjects.NewPillarID()
	domain := valueobjects.NewBusinessDomainID()

	f.addHierarchy([]capDef{
		{hr, valueobjects.LevelL1, valueobjects.CapabilityID{}},
		{recruitment, valueobjects.LevelL2, hr},
	})
	f.addRatings([]testRating{
		{hr, alwaysOn, domain, 5, "HR"},
		{hr, transform, domain, 2, "HR"},
		{recruitment, transform, domain, 4, "Recruitment"},
	})

	assertRating(t, f.resolve(t, recruitment, alwaysOn, domain), expectedRating{5, hr, true, "HR"})
	assertRating(t, f.resolve(t, recruitment, transform, domain), expectedRating{4, recruitment, false, "Recruitment"})
}
