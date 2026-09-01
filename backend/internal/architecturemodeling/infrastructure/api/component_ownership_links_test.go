package api

import (
	"testing"

	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ownershipAffordancesFor(t *testing.T, state string, role sharedctx.Role) sharedAPI.Links {
	t.Helper()
	links := NewArchitectureModelingLinks(sharedAPI.NewHATEOASLinks("/api/v1"))
	actor := sharedctx.NewActor("u1", "u@example.com", role)
	result := sharedAPI.Links{}
	links.AddOwnershipAffordances(result, "c1", state, actor)
	return result
}

func TestOwnershipAffordances_PerState(t *testing.T) {
	allAffordances := []string{"x-nominate-owner", "x-assign-owner", "x-confirm-owner", "x-clear-owner"}
	cases := []struct {
		state    string
		expected map[string]types.Link
	}{
		{
			state: valueobjects.OwnershipStateUnknown,
			expected: map[string]types.Link{
				"x-nominate-owner": {Href: "/api/v1/components/c1/ownership/nomination", Method: "POST"},
				"x-assign-owner":   {Href: "/api/v1/components/c1/ownership", Method: "PUT"},
			},
		},
		{
			state: valueobjects.OwnershipStateNominated,
			expected: map[string]types.Link{
				"x-confirm-owner": {Href: "/api/v1/components/c1/ownership/confirmation", Method: "POST"},
				"x-clear-owner":   {Href: "/api/v1/components/c1/ownership", Method: "DELETE"},
			},
		},
		{
			state: valueobjects.OwnershipStateOwned,
			expected: map[string]types.Link{
				"x-clear-owner": {Href: "/api/v1/components/c1/ownership", Method: "DELETE"},
			},
		},
		{
			state: valueobjects.OwnershipStateManaged,
			expected: map[string]types.Link{
				"x-clear-owner": {Href: "/api/v1/components/c1/ownership", Method: "DELETE"},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.state, func(t *testing.T) {
			result := ownershipAffordancesFor(t, tc.state, sharedctx.RoleArchitect)

			assertAffordances(t, result, tc.expected)
			assertNoOtherAffordances(t, result, tc.expected, allAffordances)
		})
	}
}

func assertAffordances(t *testing.T, result sharedAPI.Links, expected map[string]types.Link) {
	t.Helper()
	for rel, want := range expected {
		got, ok := result[rel]
		require.True(t, ok, "expected %s link", rel)
		assert.Equal(t, want.Method, got.Method)
		assert.Equal(t, want.Href, got.Href)
	}
}

func assertNoOtherAffordances(t *testing.T, result sharedAPI.Links, expected map[string]types.Link, all []string) {
	t.Helper()
	for _, rel := range all {
		if _, ok := expected[rel]; !ok {
			assert.NotContains(t, result, rel)
		}
	}
}

func TestOwnershipAffordances_StakeholderGetsNone(t *testing.T) {
	for _, state := range []string{
		valueobjects.OwnershipStateUnknown,
		valueobjects.OwnershipStateNominated,
		valueobjects.OwnershipStateOwned,
	} {
		result := ownershipAffordancesFor(t, state, sharedctx.RoleStakeholder)
		assert.Empty(t, result, "stakeholder must see no ownership affordances in state %s", state)
	}
}

func TestEnrichWithLinks_AddsOwnershipAffordances(t *testing.T) {
	h := &ComponentHandlers{
		hateoas: NewArchitectureModelingLinks(sharedAPI.NewHATEOASLinks("/api/v1")),
	}
	dto := &readmodels.ApplicationComponentDTO{ID: "c1", Name: "Comp", OwnershipState: valueobjects.OwnershipStateUnknown}

	h.enrichWithLinks(architectRequest(), dto)

	assert.Contains(t, dto.Links, "x-nominate-owner")
	assert.Contains(t, dto.Links, "x-assign-owner")
}

func TestStatisticsLinks_Self(t *testing.T) {
	links := NewArchitectureModelingLinks(sharedAPI.NewHATEOASLinks("/api/v1"))

	result := links.StatisticsLinks()

	self, ok := result["self"]
	require.True(t, ok)
	assert.Equal(t, "GET", self.Method)
	assert.Equal(t, "/api/v1/components/ownership-statistics", self.Href)
}
