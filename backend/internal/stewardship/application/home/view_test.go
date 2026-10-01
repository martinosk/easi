package home

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fixedNow = time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)

type fakeStore struct {
	caller   CallerFacts
	facts    PortfolioFacts
	grades   map[Subject]GradeCounts
	seeds    []PortfolioSeed
	graded   Subjects
	callerAt time.Time
	err      error
}

func (s *fakeStore) CallerFacts(_ context.Context, _ Caller, now time.Time) (CallerFacts, error) {
	s.callerAt = now
	return s.caller, s.err
}

func (s *fakeStore) Portfolio(_ context.Context, seed PortfolioSeed) (PortfolioFacts, error) {
	s.seeds = append(s.seeds, seed)
	return s.facts, nil
}

func (s *fakeStore) SubjectGrades(_ context.Context, subjects Subjects) (map[Subject]GradeCounts, error) {
	s.graded = subjects
	return s.grades, nil
}

func compose(t *testing.T, store *fakeStore, caller Caller) *Home {
	t.Helper()
	home, err := NewView(store, func() time.Time { return fixedNow }).Compose(context.Background(), caller)
	require.NoError(t, err)
	return home
}

func TestCompose_PersonalScopeCarriesTilesAndMyWork(t *testing.T) {
	store := &fakeStore{
		caller: CallerFacts{
			Stewardships: []StewardshipAnchor{{Domain: Domain{ID: "d1", Name: "Customer Engagement"}, Concern: "assessment"}},
			Ownerships: []ApplicationOwnership{
				{ApplicationID: "a1", Name: "CRM", State: "owned"},
				{ApplicationID: "a2", Name: "Portal", State: "nominated"},
			},
		},
		facts: PortfolioFacts{
			Statuses:     StatusCounts{Active: 10, Planned: 1, Deprecated: 1},
			Applications: 9,
			DomainNames:  []string{"Customer Engagement"},
			Grades:       GradeCounts{Invest: 20, Tolerate: 8, Migrate: 4, Eliminate: 2, NotAssessed: 6},
		},
		grades: map[Subject]GradeCounts{crm.subject: {Invest: 2, Eliminate: 1}},
	}

	home := compose(t, store, Caller{UserID: "u1"})

	assert.Equal(t, fixedNow, store.callerAt)
	assert.Equal(t, ScopeKindPersonal, home.Scope.Kind)
	assert.Equal(t, []ScopeStewardship{{Domain: &DomainRef{ID: "d1", Name: "Customer Engagement"}, Concern: "assessment", ConcernLabel: "Assessment"}}, home.Scope.Stewardships)
	assert.Equal(t, 2, home.Scope.OwnedApplications)
	assert.Equal(t, PortfolioSeed{DomainIDs: []string{"d1"}, CapabilityIDs: []string{}, ApplicationIDs: []string{"a1", "a2"}}, store.seeds[0])
	assert.Equal(t, &CapabilityTile{Total: 12, ByStatus: StatusBreakdown{Active: 10, Planned: 1, Deprecated: 1}}, home.Portfolio.Capabilities)
	assert.Equal(t, 9, home.Portfolio.Applications.Total)
	assert.Equal(t, &DomainTile{Total: 1, Names: []string{"Customer Engagement"}}, home.Portfolio.Domains)
	assert.Equal(t, &TimeTile{Total: 40, Shares: &TimeShares{Invest: 50, Tolerate: 20, Migrate: 10, Eliminate: 5, NotAssessed: 15}}, home.Portfolio.Time)
	assert.Equal(t, 2, home.MyWork.Total)
	assert.Equal(t, Subjects{crm.subject, portal.subject}, store.graded)
	assert.Equal(t, GradeInvest, home.MyWork.Items[0].DominantGrade)
	assert.Empty(t, home.MyWork.Items[1].DominantGrade)
}

func TestCompose_TenantScopeCoversEverythingWithEmptyMyWork(t *testing.T) {
	store := &fakeStore{facts: PortfolioFacts{Applications: 3}}

	home := compose(t, store, Caller{UserID: "u1", MayAssignStewards: true})

	assert.Equal(t, ScopeKindTenant, home.Scope.Kind)
	assert.True(t, store.seeds[0].All)
	assert.Equal(t, 3, home.Portfolio.Applications.Total)
	assert.Equal(t, &MyWork{Total: 0, Items: []WorkItem{}}, home.MyWork)
}

func TestCompose_EmptyScopeCarriesNoPortfolioAndNoMyWork(t *testing.T) {
	store := &fakeStore{}

	home := compose(t, store, Caller{UserID: "u1"})

	assert.Equal(t, ScopeKindEmpty, home.Scope.Kind)
	assert.Nil(t, home.Portfolio)
	assert.Nil(t, home.MyWork)
	assert.Empty(t, store.seeds)
}

func TestCompose_PortfolioWithoutRealisationsHasNoShares(t *testing.T) {
	store := &fakeStore{
		caller: CallerFacts{ArchitectedDomains: []Domain{{ID: "d1", Name: "Finance"}}},
		facts:  PortfolioFacts{Statuses: StatusCounts{Active: 2}},
	}

	home := compose(t, store, Caller{UserID: "u1"})

	assert.Equal(t, &TimeTile{Total: 0}, home.Portfolio.Time)
}

func TestCompose_DomainTileNamesFiveAndCountsAll(t *testing.T) {
	store := &fakeStore{
		caller: CallerFacts{ArchitectedDomains: []Domain{{ID: "d1"}}},
		facts:  PortfolioFacts{DomainNames: []string{"A", "B", "C", "D", "E", "F"}},
	}

	home := compose(t, store, Caller{UserID: "u1"})

	assert.Equal(t, &DomainTile{Total: 6, Names: []string{"A", "B", "C", "D", "E"}}, home.Portfolio.Domains)
}

func TestCompose_StoreErrorIsReturned(t *testing.T) {
	store := &fakeStore{err: errors.New("boom")}

	_, err := NewView(store, func() time.Time { return fixedNow }).Compose(context.Background(), Caller{})

	assert.Error(t, err)
}
