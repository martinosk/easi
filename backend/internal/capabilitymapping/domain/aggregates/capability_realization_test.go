package aggregates

import (
	"testing"

	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCapabilityRealization(t *testing.T) {
	capabilityID := valueobjects.NewCapabilityID()
	componentID, err := valueobjects.NewComponentIDFromString(valueobjects.NewCapabilityID().Value())
	require.NoError(t, err)

	realizationLevel, err := valueobjects.NewRealizationLevel("Full")
	require.NoError(t, err)

	notes := valueobjects.NewDescription("CRM system fully implements customer management capability")

	realization, err := NewCapabilityRealization(capabilityID, componentID, realizationLevel, notes)
	require.NoError(t, err)
	assert.NotNil(t, realization)
	assert.NotEmpty(t, realization.ID())
	assert.Equal(t, capabilityID, realization.CapabilityID())
	assert.Equal(t, componentID, realization.ComponentID())
	assert.Equal(t, realizationLevel, realization.RealizationLevel())
	assert.Equal(t, notes, realization.Notes())
	assert.NotZero(t, realization.LinkedAt())
	assert.Len(t, realization.GetUncommittedChanges(), 1)
}

func TestCapabilityRealization_RaisesLinkedEvent(t *testing.T) {
	capabilityID := valueobjects.NewCapabilityID()
	componentID, err := valueobjects.NewComponentIDFromString(valueobjects.NewCapabilityID().Value())
	require.NoError(t, err)

	realizationLevel, err := valueobjects.NewRealizationLevel("Partial")
	require.NoError(t, err)

	notes := valueobjects.NewDescription("Partially implements the capability")

	realization, err := NewCapabilityRealization(capabilityID, componentID, realizationLevel, notes)
	require.NoError(t, err)

	uncommittedEvents := realization.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "SystemLinkedToCapability", uncommittedEvents[0].EventType())
}

func TestCapabilityRealization_Update(t *testing.T) {
	capabilityID := valueobjects.NewCapabilityID()
	componentID, err := valueobjects.NewComponentIDFromString(valueobjects.NewCapabilityID().Value())
	require.NoError(t, err)

	realizationLevel, err := valueobjects.NewRealizationLevel("Planned")
	require.NoError(t, err)

	notes := valueobjects.NewDescription("Planned for Q3")

	realization, err := NewCapabilityRealization(capabilityID, componentID, realizationLevel, notes)
	require.NoError(t, err)

	realization.MarkChangesAsCommitted()

	newLevel, err := valueobjects.NewRealizationLevel("Partial")
	require.NoError(t, err)

	newNotes := valueobjects.NewDescription("Now partially implemented")

	err = realization.Update(newLevel, newNotes)
	require.NoError(t, err)

	assert.Equal(t, newLevel, realization.RealizationLevel())
	assert.Equal(t, newNotes, realization.Notes())

	uncommittedEvents := realization.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "SystemRealizationUpdated", uncommittedEvents[0].EventType())
}

func TestCapabilityRealization_Delete(t *testing.T) {
	capabilityID := valueobjects.NewCapabilityID()
	componentID, err := valueobjects.NewComponentIDFromString(valueobjects.NewCapabilityID().Value())
	require.NoError(t, err)

	realizationLevel, err := valueobjects.NewRealizationLevel("Full")
	require.NoError(t, err)

	notes := valueobjects.NewDescription("Fully implemented")

	realization, err := NewCapabilityRealization(capabilityID, componentID, realizationLevel, notes)
	require.NoError(t, err)

	realization.MarkChangesAsCommitted()

	err = realization.Delete()
	require.NoError(t, err)

	uncommittedEvents := realization.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "SystemRealizationDeleted", uncommittedEvents[0].EventType())
}

func TestCapabilityRealization_LoadFromHistory(t *testing.T) {
	capabilityID := valueobjects.NewCapabilityID()
	componentID, err := valueobjects.NewComponentIDFromString(valueobjects.NewCapabilityID().Value())
	require.NoError(t, err)

	realizationLevel, err := valueobjects.NewRealizationLevel("Full")
	require.NoError(t, err)

	notes := valueobjects.NewDescription("Test realization")

	realization, err := NewCapabilityRealization(capabilityID, componentID, realizationLevel, notes)
	require.NoError(t, err)

	events := realization.GetUncommittedChanges()

	loadedRealization, err := LoadCapabilityRealizationFromHistory(events)
	require.NoError(t, err)
	assert.NotNil(t, loadedRealization)
	assert.Equal(t, realization.ID(), loadedRealization.ID())
	assert.Equal(t, realization.CapabilityID().Value(), loadedRealization.CapabilityID().Value())
	assert.Equal(t, realization.ComponentID().Value(), loadedRealization.ComponentID().Value())
	assert.Equal(t, realization.RealizationLevel().Value(), loadedRealization.RealizationLevel().Value())
	assert.Equal(t, realization.Notes().Value(), loadedRealization.Notes().Value())
}

func TestCapabilityRealization_AllRealizationLevels(t *testing.T) {
	tests := []struct {
		name             string
		realizationLevel string
	}{
		{"Full realization", "Full"},
		{"Partial realization", "Partial"},
		{"Planned realization", "Planned"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capabilityID := valueobjects.NewCapabilityID()
			componentID, err := valueobjects.NewComponentIDFromString(valueobjects.NewCapabilityID().Value())
			require.NoError(t, err)

			realizationLevel, err := valueobjects.NewRealizationLevel(tt.realizationLevel)
			require.NoError(t, err)

			notes := valueobjects.NewDescription("Test realization")

			realization, err := NewCapabilityRealization(capabilityID, componentID, realizationLevel, notes)
			require.NoError(t, err)
			assert.NotNil(t, realization)
			assert.Equal(t, tt.realizationLevel, realization.RealizationLevel().Value())
		})
	}
}
