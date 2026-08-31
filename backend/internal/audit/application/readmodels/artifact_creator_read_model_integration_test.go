//go:build integration
// +build integration

package readmodels

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetArtifactCreators_ReturnsOneRowPerCreationEventType(t *testing.T) {
	seed, cleanup := newSeedContext(t)
	defer cleanup()

	componentID := seed.tc.uniqueID("component")
	capabilityID := seed.tc.uniqueID("capability")
	vendorID := seed.tc.uniqueID("vendor")
	teamID := seed.tc.uniqueID("team")
	entityID := seed.tc.uniqueID("entity")
	viewID := seed.tc.uniqueID("view")
	updatedComponentID := seed.tc.uniqueID("updated-component")
	domainID := seed.tc.uniqueID("domain")

	now := time.Now()
	seed.withTransaction(func(inserter eventInserter) {
		rows := []eventRow{
			{aggregateID: componentID, eventType: "ApplicationComponentCreated", data: map[string]any{"id": componentID}, version: 1, occurredAt: now, actorID: "user-1", actorEmail: "a@test.com"},
			{aggregateID: capabilityID, eventType: "CapabilityCreated", data: map[string]any{"id": capabilityID}, version: 1, occurredAt: now, actorID: "user-2", actorEmail: "b@test.com"},
			{aggregateID: vendorID, eventType: "VendorCreated", data: map[string]any{"id": vendorID}, version: 1, occurredAt: now, actorID: "user-3", actorEmail: "c@test.com"},
			{aggregateID: teamID, eventType: "InternalTeamCreated", data: map[string]any{"id": teamID}, version: 1, occurredAt: now, actorID: "user-4", actorEmail: "d@test.com"},
			{aggregateID: entityID, eventType: "AcquiredEntityCreated", data: map[string]any{"id": entityID}, version: 1, occurredAt: now, actorID: "user-5", actorEmail: "e@test.com"},
			{aggregateID: viewID, eventType: "ViewCreated", data: map[string]any{"id": viewID}, version: 1, occurredAt: now, actorID: "user-6", actorEmail: "f@test.com"},
			{aggregateID: updatedComponentID, eventType: "ApplicationComponentUpdated", data: map[string]any{"id": updatedComponentID}, version: 2, occurredAt: now, actorID: "user-7", actorEmail: "g@test.com"},
			{aggregateID: domainID, eventType: "BusinessDomainCreated", data: map[string]any{"id": domainID}, version: 1, occurredAt: now, actorID: "user-8", actorEmail: "h@test.com"},
		}
		for _, row := range rows {
			inserter.insert(row)
		}
	})

	readModel := NewArtifactCreatorReadModel(seed.tc.tenantDB)
	creators, err := readModel.GetArtifactCreators(seed.ctx)
	require.NoError(t, err)

	byAggregate := make(map[string]string, len(creators))
	for _, c := range creators {
		byAggregate[c.AggregateID] = c.CreatorID
	}

	assert.Equal(t, "user-1", byAggregate[componentID])
	assert.Equal(t, "user-2", byAggregate[capabilityID])
	assert.Equal(t, "user-3", byAggregate[vendorID])
	assert.Equal(t, "user-4", byAggregate[teamID])
	assert.Equal(t, "user-5", byAggregate[entityID])
	assert.Equal(t, "user-6", byAggregate[viewID])
	assert.NotContains(t, byAggregate, updatedComponentID)
	assert.NotContains(t, byAggregate, domainID)
}
