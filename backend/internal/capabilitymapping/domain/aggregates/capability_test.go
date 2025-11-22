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

func TestNewCapability_L2WithoutParent_OrphanAllowed(t *testing.T) {
	name, err := valueobjects.NewCapabilityName("Digital Experience")
	require.NoError(t, err)

	description := valueobjects.NewDescription("Test")

	level, err := valueobjects.NewCapabilityLevel("L2")
	require.NoError(t, err)

	var parentID valueobjects.CapabilityID

	capability, err := NewCapability(name, description, parentID, level)
	require.NoError(t, err)
	assert.NotNil(t, capability)
	assert.Equal(t, "L2", capability.Level().Value())
	assert.Empty(t, capability.ParentID().Value())
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

func TestChangeParent_L1ToL2_WhenAssignedParent(t *testing.T) {
	capability := createL1Capability(t, "Customer Engagement")
	capability.MarkChangesAsCommitted()

	newParentID := valueobjects.NewCapabilityID()
	newLevel, err := valueobjects.NewCapabilityLevel("L2")
	require.NoError(t, err)

	err = capability.ChangeParent(newParentID, newLevel)
	require.NoError(t, err)

	assert.Equal(t, newParentID.Value(), capability.ParentID().Value())
	assert.Equal(t, "L2", capability.Level().Value())

	uncommittedEvents := capability.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "CapabilityParentChanged", uncommittedEvents[0].EventType())
}

func TestChangeParent_L2ToL1_WhenOrphaned(t *testing.T) {
	capability := createL2Capability(t, "Digital Experience")
	capability.MarkChangesAsCommitted()

	var emptyParentID valueobjects.CapabilityID
	newLevel, err := valueobjects.NewCapabilityLevel("L1")
	require.NoError(t, err)

	err = capability.ChangeParent(emptyParentID, newLevel)
	require.NoError(t, err)

	assert.Empty(t, capability.ParentID().Value())
	assert.Equal(t, "L1", capability.Level().Value())
}

func TestChangeParent_L2ToL3_WhenMovedDeeperInHierarchy(t *testing.T) {
	capability := createL2Capability(t, "Digital Experience")
	capability.MarkChangesAsCommitted()

	newParentID := valueobjects.NewCapabilityID()
	newLevel, err := valueobjects.NewCapabilityLevel("L3")
	require.NoError(t, err)

	err = capability.ChangeParent(newParentID, newLevel)
	require.NoError(t, err)

	assert.Equal(t, newParentID.Value(), capability.ParentID().Value())
	assert.Equal(t, "L3", capability.Level().Value())
}

func TestChangeParent_L3ToL4_MaximumAllowedDepth(t *testing.T) {
	capability := createL3Capability(t, "Customer Portal")
	capability.MarkChangesAsCommitted()

	newParentID := valueobjects.NewCapabilityID()
	newLevel, err := valueobjects.NewCapabilityLevel("L4")
	require.NoError(t, err)

	err = capability.ChangeParent(newParentID, newLevel)
	require.NoError(t, err)

	assert.Equal(t, newParentID.Value(), capability.ParentID().Value())
	assert.Equal(t, "L4", capability.Level().Value())
}

func TestChangeParent_SelfReference_ShouldFail(t *testing.T) {
	capability := createL1Capability(t, "Customer Engagement")
	capability.MarkChangesAsCommitted()

	selfParentID, err := valueobjects.NewCapabilityIDFromString(capability.ID())
	require.NoError(t, err)

	newLevel, err := valueobjects.NewCapabilityLevel("L2")
	require.NoError(t, err)

	err = capability.ChangeParent(selfParentID, newLevel)
	assert.Error(t, err)
	assert.Equal(t, ErrCapabilityCannotBeOwnParent, err)
}

func TestChangeParent_L5CannotBeCreated_ValueObjectEnforcesMaxDepth(t *testing.T) {
	_, err := valueobjects.NewCapabilityLevel("L5")
	assert.Error(t, err)
	assert.Equal(t, valueobjects.ErrInvalidCapabilityLevel, err)
}

func TestChangeParent_ChangingParentWithinSameLevel(t *testing.T) {
	capability := createL2Capability(t, "Digital Experience")
	originalParentID := capability.ParentID()
	capability.MarkChangesAsCommitted()

	newParentID := valueobjects.NewCapabilityID()
	sameLevel, err := valueobjects.NewCapabilityLevel("L2")
	require.NoError(t, err)

	err = capability.ChangeParent(newParentID, sameLevel)
	require.NoError(t, err)

	assert.NotEqual(t, originalParentID.Value(), capability.ParentID().Value())
	assert.Equal(t, newParentID.Value(), capability.ParentID().Value())
	assert.Equal(t, "L2", capability.Level().Value())
}

func TestChangeParent_RaisesCapabilityParentChangedEvent(t *testing.T) {
	capability := createL1Capability(t, "Customer Engagement")
	originalLevel := capability.Level().Value()
	capability.MarkChangesAsCommitted()

	newParentID := valueobjects.NewCapabilityID()
	newLevel, err := valueobjects.NewCapabilityLevel("L2")
	require.NoError(t, err)

	err = capability.ChangeParent(newParentID, newLevel)
	require.NoError(t, err)

	uncommittedEvents := capability.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)

	event := uncommittedEvents[0]
	assert.Equal(t, "CapabilityParentChanged", event.EventType())

	eventData := event.EventData()
	assert.Equal(t, capability.ID(), eventData["capabilityId"])
	assert.Empty(t, eventData["oldParentId"])
	assert.Equal(t, newParentID.Value(), eventData["newParentId"])
	assert.Equal(t, originalLevel, eventData["oldLevel"])
	assert.Equal(t, "L2", eventData["newLevel"])
}

func TestChangeParent_LoadFromHistory_PreservesParentChange(t *testing.T) {
	capability := createL1Capability(t, "Customer Engagement")

	newParentID := valueobjects.NewCapabilityID()
	newLevel, err := valueobjects.NewCapabilityLevel("L2")
	require.NoError(t, err)

	err = capability.ChangeParent(newParentID, newLevel)
	require.NoError(t, err)

	allEvents := capability.GetUncommittedChanges()

	loadedCapability, err := LoadCapabilityFromHistory(allEvents)
	require.NoError(t, err)

	assert.Equal(t, capability.ID(), loadedCapability.ID())
	assert.Equal(t, newParentID.Value(), loadedCapability.ParentID().Value())
	assert.Equal(t, "L2", loadedCapability.Level().Value())
}

func TestChangeParent_MultipleParentChanges(t *testing.T) {
	capability := createL1Capability(t, "Customer Engagement")
	capability.MarkChangesAsCommitted()

	firstParentID := valueobjects.NewCapabilityID()
	levelL2, err := valueobjects.NewCapabilityLevel("L2")
	require.NoError(t, err)

	err = capability.ChangeParent(firstParentID, levelL2)
	require.NoError(t, err)
	capability.MarkChangesAsCommitted()

	secondParentID := valueobjects.NewCapabilityID()
	levelL3, err := valueobjects.NewCapabilityLevel("L3")
	require.NoError(t, err)

	err = capability.ChangeParent(secondParentID, levelL3)
	require.NoError(t, err)

	assert.Equal(t, secondParentID.Value(), capability.ParentID().Value())
	assert.Equal(t, "L3", capability.Level().Value())
}

func TestChangeParent_L4ToL1_Orphaning(t *testing.T) {
	capability := createL4Capability(t, "Deep Feature")
	capability.MarkChangesAsCommitted()

	var emptyParentID valueobjects.CapabilityID
	levelL1, err := valueobjects.NewCapabilityLevel("L1")
	require.NoError(t, err)

	err = capability.ChangeParent(emptyParentID, levelL1)
	require.NoError(t, err)

	assert.Empty(t, capability.ParentID().Value())
	assert.Equal(t, "L1", capability.Level().Value())
}

func TestChangeParent_PreservesOtherAggregateState(t *testing.T) {
	name, err := valueobjects.NewCapabilityName("Customer Engagement")
	require.NoError(t, err)

	description := valueobjects.NewDescription("Customer-facing capabilities")

	level, err := valueobjects.NewCapabilityLevel("L1")
	require.NoError(t, err)

	var parentID valueobjects.CapabilityID

	capability, err := NewCapability(name, description, parentID, level)
	require.NoError(t, err)
	capability.MarkChangesAsCommitted()

	newParentID := valueobjects.NewCapabilityID()
	newLevel, err := valueobjects.NewCapabilityLevel("L2")
	require.NoError(t, err)

	err = capability.ChangeParent(newParentID, newLevel)
	require.NoError(t, err)

	assert.Equal(t, name.Value(), capability.Name().Value())
	assert.Equal(t, description.Value(), capability.Description().Value())
}

func createL1Capability(t *testing.T, capabilityName string) *Capability {
	t.Helper()

	name, err := valueobjects.NewCapabilityName(capabilityName)
	require.NoError(t, err)

	description := valueobjects.NewDescription("Test capability")

	level, err := valueobjects.NewCapabilityLevel("L1")
	require.NoError(t, err)

	var parentID valueobjects.CapabilityID

	capability, err := NewCapability(name, description, parentID, level)
	require.NoError(t, err)

	return capability
}

func createL2Capability(t *testing.T, capabilityName string) *Capability {
	t.Helper()

	name, err := valueobjects.NewCapabilityName(capabilityName)
	require.NoError(t, err)

	description := valueobjects.NewDescription("Test capability")

	level, err := valueobjects.NewCapabilityLevel("L2")
	require.NoError(t, err)

	parentID := valueobjects.NewCapabilityID()

	capability, err := NewCapability(name, description, parentID, level)
	require.NoError(t, err)

	return capability
}

func createL3Capability(t *testing.T, capabilityName string) *Capability {
	t.Helper()

	name, err := valueobjects.NewCapabilityName(capabilityName)
	require.NoError(t, err)

	description := valueobjects.NewDescription("Test capability")

	level, err := valueobjects.NewCapabilityLevel("L3")
	require.NoError(t, err)

	parentID := valueobjects.NewCapabilityID()

	capability, err := NewCapability(name, description, parentID, level)
	require.NoError(t, err)

	return capability
}

func createL4Capability(t *testing.T, capabilityName string) *Capability {
	t.Helper()

	name, err := valueobjects.NewCapabilityName(capabilityName)
	require.NoError(t, err)

	description := valueobjects.NewDescription("Test capability")

	level, err := valueobjects.NewCapabilityLevel("L4")
	require.NoError(t, err)

	parentID := valueobjects.NewCapabilityID()

	capability, err := NewCapability(name, description, parentID, level)
	require.NoError(t, err)

	return capability
}

