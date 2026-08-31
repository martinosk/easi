package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/architecturedirection/application/readmodels"
)

type fakeCapabilityNodes struct {
	nodes map[string]readmodels.CapabilityNodeDTO
}

func (f fakeCapabilityNodes) GetByID(_ context.Context, capabilityID string) (*readmodels.CapabilityNodeDTO, error) {
	node, ok := f.nodes[capabilityID]
	if !ok {
		return nil, nil
	}
	return &node, nil
}

type fakeReferenceCache struct {
	present map[string]bool
}

func (f fakeReferenceCache) Exists(_ context.Context, entity readmodels.ReferenceEntity, entityID string) (bool, error) {
	return f.present[string(entity)+"|"+entityID], nil
}

type fakeDirectRealizations struct {
	byPair map[string]string
}

func (f fakeDirectRealizations) DirectRealizationID(_ context.Context, capabilityID readmodels.CapabilityID, componentID readmodels.ComponentID) (readmodels.RealizationID, bool, error) {
	id, ok := f.byPair[string(capabilityID)+"|"+string(componentID)]
	return readmodels.RealizationID(id), ok, nil
}

func testReferenceLookups() referenceLookups {
	return newReferenceLookups(
		fakeCapabilityNodes{nodes: map[string]readmodels.CapabilityNodeDTO{
			"cap-1": {CapabilityID: "cap-1", BusinessDomainID: "bd-1"},
			"cap-2": {CapabilityID: "cap-2"},
		}},
		fakeReferenceCache{present: map[string]bool{
			"application|comp-1":     true,
			"business_domain|bd-1":   true,
			"application|bd-1":       false,
			"business_domain|comp-1": false,
		}},
		fakeDirectRealizations{byPair: map[string]string{"cap-1|comp-1": "r-1"}},
	)
}

func TestReferenceLookups_CapabilityExistenceComesFromTheNodeCache(t *testing.T) {
	lookups := testReferenceLookups()

	exists, err := lookups.capabilityExists(context.Background(), "cap-1")
	require.NoError(t, err)
	assert.True(t, exists)

	missing, err := lookups.capabilityExists(context.Background(), "cap-unknown")
	require.NoError(t, err)
	assert.False(t, missing)
}

func TestReferenceLookups_ComponentAndDomainExistenceAreScopedToTheirKind(t *testing.T) {
	lookups := testReferenceLookups()

	component, err := lookups.componentExists(context.Background(), "comp-1")
	require.NoError(t, err)
	assert.True(t, component)

	componentOfDomainID, err := lookups.componentExists(context.Background(), "bd-1")
	require.NoError(t, err)
	assert.False(t, componentOfDomainID)

	domain, err := lookups.domainExists(context.Background(), "bd-1")
	require.NoError(t, err)
	assert.True(t, domain)

	domainOfComponentID, err := lookups.domainExists(context.Background(), "comp-1")
	require.NoError(t, err)
	assert.False(t, domainOfComponentID)
}

func TestReferenceLookups_EffectiveDomainComesFromTheNodeCache(t *testing.T) {
	lookups := testReferenceLookups()

	inDomain, err := lookups.capabilityEffectivelyInDomain(context.Background(), "cap-1", "bd-1")
	require.NoError(t, err)
	assert.True(t, inDomain)

	otherDomain, err := lookups.capabilityEffectivelyInDomain(context.Background(), "cap-1", "bd-2")
	require.NoError(t, err)
	assert.False(t, otherDomain)

	unassigned, err := lookups.capabilityEffectivelyInDomain(context.Background(), "cap-2", "bd-1")
	require.NoError(t, err)
	assert.False(t, unassigned)

	unknown, err := lookups.capabilityEffectivelyInDomain(context.Background(), "cap-unknown", "bd-1")
	require.NoError(t, err)
	assert.False(t, unknown)
}

func TestReferenceLookups_DirectRealizationComesFromTheRealizationCache(t *testing.T) {
	lookups := testReferenceLookups()

	realizationID, found, err := lookups.directRealization(context.Background(), "cap-1", "comp-1")
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, "r-1", realizationID)

	_, missing, err := lookups.directRealization(context.Background(), "cap-2", "comp-1")
	require.NoError(t, err)
	assert.False(t, missing)
}
