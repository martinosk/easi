package aggregates

import (
	"testing"

	"easi/backend/internal/capabilitymapping/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetStrategyImportance(t *testing.T) {
	params := newImportanceParams(t)

	aggregate, err := SetStrategyImportance(params)
	require.NoError(t, err)
	assert.NotNil(t, aggregate)
	assert.NotEmpty(t, aggregate.ID())
	assert.Equal(t, params.BusinessDomainID, aggregate.BusinessDomainID())
	assert.Equal(t, params.CapabilityID, aggregate.CapabilityID())
	assert.Equal(t, params.PillarID, aggregate.PillarID())
	assert.Equal(t, params.Importance, aggregate.Importance())
	assert.Equal(t, params.Rationale, aggregate.Rationale())
	assert.Len(t, aggregate.GetUncommittedChanges(), 1)
}

func TestSetStrategyImportance_WithEmptyRationale(t *testing.T) {
	params := newImportanceParams(t)
	params.Rationale = valueobjects.EmptyRationale()

	aggregate, err := SetStrategyImportance(params)
	require.NoError(t, err)
	assert.True(t, aggregate.Rationale().IsEmpty())
}

func TestSetStrategyImportance_RaisesEvent(t *testing.T) {
	params := newImportanceParams(t)

	aggregate, err := SetStrategyImportance(params)
	require.NoError(t, err)

	events := aggregate.GetUncommittedChanges()
	require.Len(t, events, 1)
	assert.Equal(t, "StrategyImportanceSet", events[0].EventType())

	eventData := events[0].EventData()
	assert.Equal(t, aggregate.ID(), eventData["id"])
	assert.Equal(t, params.BusinessDomainID.Value(), eventData["businessDomainId"])
	assert.Equal(t, params.CapabilityID.Value(), eventData["capabilityId"])
	assert.Equal(t, params.PillarID.Value(), eventData["pillarId"])
	assert.Equal(t, params.PillarName, eventData["pillarName"])
	assert.Equal(t, params.Importance.Value(), eventData["importance"])
	assert.Equal(t, params.Rationale.Value(), eventData["rationale"])
}

func TestUpdateStrategyImportance(t *testing.T) {
	aggregate := createStrategyImportance(t)
	aggregate.MarkChangesAsCommitted()

	newImportance, _ := valueobjects.NewImportance(5)
	newRationale, _ := valueobjects.NewRationale("Now critical for success")

	err := aggregate.Update(newImportance, newRationale)
	require.NoError(t, err)
	assert.Equal(t, newImportance, aggregate.Importance())
	assert.Equal(t, newRationale, aggregate.Rationale())

	events := aggregate.GetUncommittedChanges()
	require.Len(t, events, 1)
	assert.Equal(t, "StrategyImportanceUpdated", events[0].EventType())
}

func TestUpdateStrategyImportance_EventContainsOldValues(t *testing.T) {
	aggregate := createStrategyImportance(t)
	oldImportance := aggregate.Importance().Value()
	oldRationale := aggregate.Rationale().Value()
	aggregate.MarkChangesAsCommitted()

	newImportance, _ := valueobjects.NewImportance(5)
	newRationale, _ := valueobjects.NewRationale("Updated rationale")

	err := aggregate.Update(newImportance, newRationale)
	require.NoError(t, err)

	events := aggregate.GetUncommittedChanges()
	eventData := events[0].EventData()
	assert.Equal(t, oldImportance, eventData["oldImportance"])
	assert.Equal(t, oldRationale, eventData["oldRationale"])
}

func TestRemoveStrategyImportance(t *testing.T) {
	aggregate := createStrategyImportance(t)
	aggregate.MarkChangesAsCommitted()

	err := aggregate.Remove()
	require.NoError(t, err)

	events := aggregate.GetUncommittedChanges()
	require.Len(t, events, 1)
	assert.Equal(t, "StrategyImportanceRemoved", events[0].EventType())

	eventData := events[0].EventData()
	assert.Equal(t, aggregate.ID(), eventData["id"])
	assert.Equal(t, aggregate.BusinessDomainID().Value(), eventData["businessDomainId"])
	assert.Equal(t, aggregate.CapabilityID().Value(), eventData["capabilityId"])
	assert.Equal(t, aggregate.PillarID().Value(), eventData["pillarId"])
}

func TestLoadStrategyImportanceFromHistory(t *testing.T) {
	original := createStrategyImportance(t)
	events := original.GetUncommittedChanges()

	loaded, err := LoadStrategyImportanceFromHistory(events)
	require.NoError(t, err)
	assert.Equal(t, original.ID(), loaded.ID())
	assert.Equal(t, original.BusinessDomainID().Value(), loaded.BusinessDomainID().Value())
	assert.Equal(t, original.CapabilityID().Value(), loaded.CapabilityID().Value())
	assert.Equal(t, original.PillarID().Value(), loaded.PillarID().Value())
	assert.Equal(t, original.Importance().Value(), loaded.Importance().Value())
	assert.Equal(t, original.Rationale().Value(), loaded.Rationale().Value())
}

func TestLoadStrategyImportanceFromHistory_WithUpdate(t *testing.T) {
	aggregate := createStrategyImportance(t)

	newImportance, _ := valueobjects.NewImportance(5)
	newRationale, _ := valueobjects.NewRationale("Updated")
	_ = aggregate.Update(newImportance, newRationale)

	events := aggregate.GetUncommittedChanges()

	loaded, err := LoadStrategyImportanceFromHistory(events)
	require.NoError(t, err)
	assert.Equal(t, 5, loaded.Importance().Value())
	assert.Equal(t, "Updated", loaded.Rationale().Value())
}

func newImportanceParams(t *testing.T) NewImportanceParams {
	t.Helper()

	businessDomainID, _ := valueobjects.NewBusinessDomainIDFromString("550e8400-e29b-41d4-a716-446655440000")
	capabilityID, _ := valueobjects.NewCapabilityIDFromString("550e8400-e29b-41d4-a716-446655440001")
	pillarID, _ := valueobjects.NewPillarIDFromString("550e8400-e29b-41d4-a716-446655440002")
	importance, _ := valueobjects.NewImportance(3)
	rationale, _ := valueobjects.NewRationale("Initial rationale")

	return NewImportanceParams{
		BusinessDomainID: businessDomainID,
		CapabilityID:     capabilityID,
		PillarID:         pillarID,
		PillarName:       "Test Pillar",
		Importance:       importance,
		Rationale:        rationale,
	}
}

func createStrategyImportance(t *testing.T) *StrategyImportance {
	t.Helper()

	aggregate, err := SetStrategyImportance(newImportanceParams(t))
	require.NoError(t, err)

	return aggregate
}
