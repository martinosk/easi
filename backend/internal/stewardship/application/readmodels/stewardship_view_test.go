package readmodels

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeDomainSource map[string]*cachedDomainWithArchitect

func (f fakeDomainSource) getWithArchitect(_ context.Context, domainID string) (*cachedDomainWithArchitect, error) {
	return f[domainID], nil
}

type fakeRowSource []StewardshipRow

func (f fakeRowSource) ForDomain(_ context.Context, _ string) ([]StewardshipRow, error) {
	return f, nil
}

func strPtr(s string) *string { return &s }

func TestDomainStewardships_ListsEveryConcernAssignedOrNot(t *testing.T) {
	assignedAt := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	view := newStewardshipView(
		fakeDomainSource{"ce": {ID: "ce", Name: "Customer Engagement", DomainArchitectID: strPtr("alice"), DomainArchitectName: strPtr("Alice Smith")}},
		fakeRowSource{{ID: "s-1", DomainID: "ce", Concern: "ownership", StewardID: "mette", StewardName: strPtr("Mette Gram"), AssignedBy: "bob@example.com", AssignedAt: assignedAt}},
	)

	result, err := view.DomainStewardships(context.Background(), "ce")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "ce", result.DomainID)
	assert.Equal(t, "Customer Engagement", result.DomainName)
	require.NotNil(t, result.Fallback)
	assert.Equal(t, "alice", result.Fallback.ID)
	assert.Equal(t, "Alice Smith", *result.Fallback.Name)

	var concerns []string
	for _, c := range result.Concerns {
		concerns = append(concerns, c.Concern)
		assert.NotEmpty(t, c.Label)
		assert.NotEmpty(t, c.Description)
	}
	assert.Equal(t, []string{"ownership", "assessment", "documentation", "planning", "structure"}, concerns)

	ownership := result.Concerns[0]
	require.NotNil(t, ownership.Steward)
	assert.Equal(t, "mette", ownership.Steward.ID)
	assert.Equal(t, "Mette Gram", *ownership.Steward.Name)
	assert.Equal(t, "bob@example.com", *ownership.AssignedBy)
	assert.Equal(t, assignedAt, *ownership.AssignedAt)
	for _, unassigned := range result.Concerns[1:] {
		assert.Nil(t, unassigned.Steward, unassigned.Concern)
		assert.Nil(t, unassigned.AssignedBy)
		assert.Nil(t, unassigned.AssignedAt)
	}
}

func TestDomainStewardships_NoFallbackWithoutADomainArchitect(t *testing.T) {
	view := newStewardshipView(fakeDomainSource{"lg": {ID: "lg", Name: "Logistics"}}, fakeRowSource{})

	result, err := view.DomainStewardships(context.Background(), "lg")

	require.NoError(t, err)
	assert.Nil(t, result.Fallback)
	assert.Len(t, result.Concerns, 5)
}

func TestDomainStewardships_FallbackWithUnknownNameKeepsTheID(t *testing.T) {
	view := newStewardshipView(fakeDomainSource{"lg": {ID: "lg", Name: "Logistics", DomainArchitectID: strPtr("ghost")}}, fakeRowSource{})

	result, err := view.DomainStewardships(context.Background(), "lg")

	require.NoError(t, err)
	require.NotNil(t, result.Fallback)
	assert.Equal(t, "ghost", result.Fallback.ID)
	assert.Nil(t, result.Fallback.Name)
}

func TestDomainStewardships_UnknownDomainIsNil(t *testing.T) {
	view := newStewardshipView(fakeDomainSource{}, fakeRowSource{})

	result, err := view.DomainStewardships(context.Background(), "nope")

	require.NoError(t, err)
	assert.Nil(t, result)
}
