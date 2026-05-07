package aggregates

import (
	"testing"

	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newEnterpriseImportanceParams(t *testing.T) NewEnterpriseImportanceParams {
	t.Helper()

	importance, _ := valueobjects.NewImportance(3)
	rationale, _ := valueobjects.NewRationale("Test rationale")

	return NewEnterpriseImportanceParams{
		EnterpriseCapabilityID: valueobjects.NewEnterpriseCapabilityID(),
		PillarID:               valueobjects.NewPillarID(),
		PillarName:             "Test Pillar",
		Importance:             importance,
		Rationale:              rationale,
	}
}

func TestSetEnterpriseStrategicImportance(t *testing.T) {
	params := newEnterpriseImportanceParams(t)
	params.PillarName = "Standardization"
	importance, _ := valueobjects.NewImportance(5)
	params.Importance = importance
	rationale, _ := valueobjects.NewRationale("Critical for business operations")
	params.Rationale = rationale

	si, err := SetEnterpriseStrategicImportance(params)
	require.NoError(t, err)

	assert.NotNil(t, si)
	assert.NotEmpty(t, si.ID())
	assert.True(t, params.EnterpriseCapabilityID.Equals(si.EnterpriseCapabilityID()))
	assert.True(t, params.PillarID.Equals(si.PillarID()))
	assert.True(t, params.Importance.Equals(si.Importance()))
	assert.True(t, params.Rationale.Equals(si.Rationale()))
	assert.False(t, si.SetAt().IsZero())
	assert.Len(t, si.GetUncommittedChanges(), 1)
}

func TestSetEnterpriseStrategicImportance_DeterministicID(t *testing.T) {
	params := newEnterpriseImportanceParams(t)

	si1, _ := SetEnterpriseStrategicImportance(params)
	si2, _ := SetEnterpriseStrategicImportance(params)

	assert.Equal(t, si1.ID(), si2.ID())
}

func TestEnterpriseStrategicImportance_RaisesSetEvent(t *testing.T) {
	params := newEnterpriseImportanceParams(t)
	params.PillarName = "Innovation"
	importance, _ := valueobjects.NewImportance(4)
	params.Importance = importance
	rationale, _ := valueobjects.NewRationale("Important for strategy")
	params.Rationale = rationale

	si, _ := SetEnterpriseStrategicImportance(params)

	uncommittedEvents := si.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "EnterpriseStrategicImportanceSet", uncommittedEvents[0].EventType())

	eventData := uncommittedEvents[0].EventData()
	assert.Equal(t, si.ID(), eventData["id"])
	assert.Equal(t, params.EnterpriseCapabilityID.Value(), eventData["enterpriseCapabilityId"])
	assert.Equal(t, params.PillarID.Value(), eventData["pillarId"])
	assert.Equal(t, "Innovation", eventData["pillarName"])
	assert.Equal(t, 4, eventData["importance"])
	assert.Equal(t, "Important for strategy", eventData["rationale"])
}

func TestEnterpriseStrategicImportance_Update(t *testing.T) {
	si := createEnterpriseStrategicImportance(t)
	si.MarkChangesAsCommitted()

	newImportance, _ := valueobjects.NewImportance(5)
	newRationale, _ := valueobjects.NewRationale("Updated rationale")

	err := si.Update(newImportance, newRationale)
	require.NoError(t, err)

	assert.True(t, newImportance.Equals(si.Importance()))
	assert.True(t, newRationale.Equals(si.Rationale()))

	uncommittedEvents := si.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "EnterpriseStrategicImportanceUpdated", uncommittedEvents[0].EventType())
}

func TestEnterpriseStrategicImportance_Remove(t *testing.T) {
	si := createEnterpriseStrategicImportance(t)
	si.MarkChangesAsCommitted()

	err := si.Remove()
	require.NoError(t, err)

	uncommittedEvents := si.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "EnterpriseStrategicImportanceRemoved", uncommittedEvents[0].EventType())

	eventData := uncommittedEvents[0].EventData()
	assert.Equal(t, si.ID(), eventData["id"])
	assert.Equal(t, si.EnterpriseCapabilityID().Value(), eventData["enterpriseCapabilityId"])
	assert.Equal(t, si.PillarID().Value(), eventData["pillarId"])
}

func TestEnterpriseStrategicImportance_LoadFromHistory(t *testing.T) {
	params := newEnterpriseImportanceParams(t)

	si, _ := SetEnterpriseStrategicImportance(params)

	events := si.GetUncommittedChanges()

	loadedSI, err := LoadEnterpriseStrategicImportanceFromHistory(events)
	require.NoError(t, err)

	assert.Equal(t, si.ID(), loadedSI.ID())
	assert.Equal(t, params.EnterpriseCapabilityID.Value(), loadedSI.EnterpriseCapabilityID().Value())
	assert.Equal(t, params.PillarID.Value(), loadedSI.PillarID().Value())
	assert.Equal(t, params.Importance.Value(), loadedSI.Importance().Value())
	assert.Equal(t, params.Rationale.Value(), loadedSI.Rationale().Value())
}

func TestEnterpriseStrategicImportance_LoadFromHistoryWithUpdate(t *testing.T) {
	si := createEnterpriseStrategicImportance(t)

	newImportance, _ := valueobjects.NewImportance(5)
	newRationale, _ := valueobjects.NewRationale("Updated")

	_ = si.Update(newImportance, newRationale)

	events := si.GetUncommittedChanges()

	loadedSI, err := LoadEnterpriseStrategicImportanceFromHistory(events)
	require.NoError(t, err)

	assert.Equal(t, newImportance.Value(), loadedSI.Importance().Value())
	assert.Equal(t, newRationale.Value(), loadedSI.Rationale().Value())
}

func TestEnterpriseStrategicImportance_RemoveMultipleTimes(t *testing.T) {
	si := createEnterpriseStrategicImportance(t)
	si.MarkChangesAsCommitted()

	err := si.Remove()
	require.NoError(t, err)

	err = si.Remove()
	require.NoError(t, err)

	uncommittedEvents := si.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 2)
}

func TestEnterpriseStrategicImportance_UpdateMultipleTimes(t *testing.T) {
	si := createEnterpriseStrategicImportance(t)
	si.MarkChangesAsCommitted()

	for i := 1; i <= 3; i++ {
		newImportance, _ := valueobjects.NewImportance((i % 5) + 1)
		newRationale, _ := valueobjects.NewRationale("Rationale " + string(rune('A'+i)))

		err := si.Update(newImportance, newRationale)
		require.NoError(t, err)
	}

	uncommittedEvents := si.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 3)
}

func TestEnterpriseStrategicImportance_LoadFromEmptyHistory(t *testing.T) {
	loadedSI, err := LoadEnterpriseStrategicImportanceFromHistory(nil)
	require.NoError(t, err)

	assert.NotNil(t, loadedSI)
	assert.NotEmpty(t, loadedSI.ID())
	assert.Empty(t, loadedSI.GetUncommittedChanges())
}

func TestEnterpriseStrategicImportance_LoadFromHistoryWithRemove(t *testing.T) {
	si := createEnterpriseStrategicImportance(t)
	_ = si.Remove()

	events := si.GetUncommittedChanges()

	loadedSI, err := LoadEnterpriseStrategicImportanceFromHistory(events)
	require.NoError(t, err)

	assert.Equal(t, si.ID(), loadedSI.ID())
}

func createEnterpriseStrategicImportance(t *testing.T) *EnterpriseStrategicImportance {
	t.Helper()

	si, err := SetEnterpriseStrategicImportance(newEnterpriseImportanceParams(t))
	require.NoError(t, err)

	return si
}
