package aggregates

import (
	"testing"

	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCapability_L1(t *testing.T) {
	name, err := valueobjects.NewCapabilityName("Customer Engagement")
	require.NoError(t, err)

	description := valueobjects.NewDescription("All customer-facing capabilities")

	level, err := valueobjects.NewCapabilityLevel("L1")
	require.NoError(t, err)

	var parentID valueobjects.CapabilityID

	capability, err := NewCapability(name, description, parentID, level)
	require.NoError(t, err)
	assert.NotNil(t, capability)
	assert.NotEmpty(t, capability.ID())
	assert.Equal(t, name, capability.Name())
	assert.Equal(t, description, capability.Description())
	assert.Equal(t, level, capability.Level())
	assert.Empty(t, capability.ParentID().Value())
	assert.NotZero(t, capability.CreatedAt())
	assert.Len(t, capability.GetUncommittedChanges(), 1)
}

func TestNewCapability_L1WithParent_ShouldFail(t *testing.T) {
	name, err := valueobjects.NewCapabilityName("Customer Engagement")
	require.NoError(t, err)

	description := valueobjects.NewDescription("Test")

	level, err := valueobjects.NewCapabilityLevel("L1")
	require.NoError(t, err)

	parentID := valueobjects.NewCapabilityID()

	capability, err := NewCapability(name, description, parentID, level)
	assert.Error(t, err)
	assert.Nil(t, capability)
	assert.Equal(t, ErrL1CannotHaveParent, err)
}

func TestNewCapability_L2WithoutParent_ShouldFail(t *testing.T) {
	name, err := valueobjects.NewCapabilityName("Digital Experience")
	require.NoError(t, err)

	description := valueobjects.NewDescription("Test")

	level, err := valueobjects.NewCapabilityLevel("L2")
	require.NoError(t, err)

	var parentID valueobjects.CapabilityID

	capability, err := NewCapability(name, description, parentID, level)
	assert.Error(t, err)
	assert.Nil(t, capability)
	assert.Equal(t, ErrNonL1MustHaveParent, err)
}

func TestNewCapability_L2WithParent(t *testing.T) {
	name, err := valueobjects.NewCapabilityName("Digital Experience")
	require.NoError(t, err)

	description := valueobjects.NewDescription("Customer digital touchpoints")

	level, err := valueobjects.NewCapabilityLevel("L2")
	require.NoError(t, err)

	parentID := valueobjects.NewCapabilityID()

	capability, err := NewCapability(name, description, parentID, level)
	require.NoError(t, err)
	assert.NotNil(t, capability)
	assert.Equal(t, parentID.Value(), capability.ParentID().Value())
}

func TestCapability_RaisesCreatedEvent(t *testing.T) {
	name, err := valueobjects.NewCapabilityName("Operations")
	require.NoError(t, err)

	description := valueobjects.NewDescription("Test description")

	level, err := valueobjects.NewCapabilityLevel("L1")
	require.NoError(t, err)

	var parentID valueobjects.CapabilityID

	capability, err := NewCapability(name, description, parentID, level)
	require.NoError(t, err)

	uncommittedEvents := capability.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "CapabilityCreated", uncommittedEvents[0].EventType())
}

func TestCapability_Update(t *testing.T) {
	name, err := valueobjects.NewCapabilityName("Finance")
	require.NoError(t, err)

	description := valueobjects.NewDescription("Financial capabilities")

	level, err := valueobjects.NewCapabilityLevel("L1")
	require.NoError(t, err)

	var parentID valueobjects.CapabilityID

	capability, err := NewCapability(name, description, parentID, level)
	require.NoError(t, err)

	capability.MarkChangesAsCommitted()

	newName, err := valueobjects.NewCapabilityName("Finance & Accounting")
	require.NoError(t, err)

	newDescription := valueobjects.NewDescription("Financial and accounting capabilities")

	err = capability.Update(newName, newDescription)
	require.NoError(t, err)

	assert.Equal(t, newName, capability.Name())
	assert.Equal(t, newDescription, capability.Description())

	uncommittedEvents := capability.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "CapabilityUpdated", uncommittedEvents[0].EventType())
}

func TestCapability_LoadFromHistory(t *testing.T) {
	name, err := valueobjects.NewCapabilityName("IT Infrastructure")
	require.NoError(t, err)

	description := valueobjects.NewDescription("Infrastructure capabilities")

	level, err := valueobjects.NewCapabilityLevel("L1")
	require.NoError(t, err)

	var parentID valueobjects.CapabilityID

	capability, err := NewCapability(name, description, parentID, level)
	require.NoError(t, err)

	events := capability.GetUncommittedChanges()

	loadedCapability, err := LoadCapabilityFromHistory(events)
	require.NoError(t, err)
	assert.NotNil(t, loadedCapability)
	assert.Equal(t, capability.ID(), loadedCapability.ID())
	assert.Equal(t, capability.Name().Value(), loadedCapability.Name().Value())
	assert.Equal(t, capability.Description().Value(), loadedCapability.Description().Value())
	assert.Equal(t, capability.Level().Value(), loadedCapability.Level().Value())
}
