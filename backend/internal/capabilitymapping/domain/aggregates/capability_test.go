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

	description := valueobjects.MustNewDescription("All customer-facing capabilities")

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
	name, level := newCapabilityInputs(t, "Customer Engagement", "L1")
	description := valueobjects.MustNewDescription("Test")
	parentID := valueobjects.NewCapabilityID()

	capability, err := NewCapability(name, description, parentID, level)
	assert.Error(t, err)
	assert.Nil(t, capability)
	assert.Equal(t, ErrL1CannotHaveParent, err)
}

func TestNewCapability_L2WithoutParent_OrphanAllowed(t *testing.T) {
	name, level := newCapabilityInputs(t, "Digital Experience", "L2")
	description := valueobjects.MustNewDescription("Test")
	var parentID valueobjects.CapabilityID

	capability, err := NewCapability(name, description, parentID, level)
	require.NoError(t, err)
	assert.NotNil(t, capability)
	assert.Equal(t, "L2", capability.Level().Value())
	assert.Empty(t, capability.ParentID().Value())
}

func TestNewCapability_L2WithParent(t *testing.T) {
	name, level := newCapabilityInputs(t, "Digital Experience", "L2")
	description := valueobjects.MustNewDescription("Customer digital touchpoints")
	parentID := valueobjects.NewCapabilityID()

	capability, err := NewCapability(name, description, parentID, level)
	require.NoError(t, err)
	assert.NotNil(t, capability)
	assert.Equal(t, parentID.Value(), capability.ParentID().Value())
}

func newCapabilityInputs(t *testing.T, capName string, levelStr string) (valueobjects.CapabilityName, valueobjects.CapabilityLevel) {
	t.Helper()

	name, err := valueobjects.NewCapabilityName(capName)
	require.NoError(t, err)

	level, err := valueobjects.NewCapabilityLevel(levelStr)
	require.NoError(t, err)

	return name, level
}

func TestCapability_RaisesCreatedEvent(t *testing.T) {
	name, err := valueobjects.NewCapabilityName("Operations")
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Test description")

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

	description := valueobjects.MustNewDescription("Financial capabilities")

	level, err := valueobjects.NewCapabilityLevel("L1")
	require.NoError(t, err)

	var parentID valueobjects.CapabilityID

	capability, err := NewCapability(name, description, parentID, level)
	require.NoError(t, err)

	capability.MarkChangesAsCommitted()

	newName, err := valueobjects.NewCapabilityName("Finance & Accounting")
	require.NoError(t, err)

	newDescription := valueobjects.MustNewDescription("Financial and accounting capabilities")

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

	description := valueobjects.MustNewDescription("Infrastructure capabilities")

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
	capability := createCapability(t, "Customer Engagement", "L1")
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

func TestChangeParent_LevelTransitions(t *testing.T) {
	tests := []struct {
		name         string
		fromLevel    string
		toLevel      string
		capName      string
		orphan       bool
	}{
		{"L2ToL1_WhenOrphaned", "L2", "L1", "Digital Experience", true},
		{"L2ToL3_WhenMovedDeeperInHierarchy", "L2", "L3", "Digital Experience", false},
		{"L3ToL4_MaximumAllowedDepth", "L3", "L4", "Customer Portal", false},
		{"L4ToL1_Orphaning", "L4", "L1", "Deep Feature", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capability := createCapability(t, tt.capName, tt.fromLevel)
			capability.MarkChangesAsCommitted()

			var newParentID valueobjects.CapabilityID
			if !tt.orphan {
				newParentID = valueobjects.NewCapabilityID()
			}
			newLevel, err := valueobjects.NewCapabilityLevel(tt.toLevel)
			require.NoError(t, err)

			err = capability.ChangeParent(newParentID, newLevel)
			require.NoError(t, err)

			if tt.orphan {
				assert.Empty(t, capability.ParentID().Value())
			} else {
				assert.Equal(t, newParentID.Value(), capability.ParentID().Value())
			}
			assert.Equal(t, tt.toLevel, capability.Level().Value())
		})
	}
}

func TestChangeParent_SelfReference_ShouldFail(t *testing.T) {
	capability := createCapability(t, "Customer Engagement", "L1")
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
	capability := createCapability(t, "Digital Experience", "L2")
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
	capability := createCapability(t, "Customer Engagement", "L1")
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
	capability := createCapability(t, "Customer Engagement", "L1")

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
	capability := createCapability(t, "Customer Engagement", "L1")
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

func TestChangeParent_PreservesOtherAggregateState(t *testing.T) {
	capability := createCapability(t, "Customer Engagement", "L1")
	originalName := capability.Name().Value()
	originalDescription := capability.Description().Value()
	capability.MarkChangesAsCommitted()

	newParentID := valueobjects.NewCapabilityID()
	newLevel, err := valueobjects.NewCapabilityLevel("L2")
	require.NoError(t, err)

	err = capability.ChangeParent(newParentID, newLevel)
	require.NoError(t, err)

	assert.Equal(t, originalName, capability.Name().Value())
	assert.Equal(t, originalDescription, capability.Description().Value())
}

func createCapability(t *testing.T, capabilityName string, levelStr string) *Capability {
	t.Helper()

	name, err := valueobjects.NewCapabilityName(capabilityName)
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Test capability")

	level, err := valueobjects.NewCapabilityLevel(levelStr)
	require.NoError(t, err)

	var parentID valueobjects.CapabilityID
	if levelStr != "L1" {
		parentID = valueobjects.NewCapabilityID()
	}

	capability, err := NewCapability(name, description, parentID, level)
	require.NoError(t, err)

	return capability
}

func TestCapability_Delete_RaisesDeletedEvent(t *testing.T) {
	capability := createCapability(t, "Customer Engagement", "L1")
	capability.MarkChangesAsCommitted()

	err := capability.Delete()
	require.NoError(t, err)

	uncommittedEvents := capability.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "CapabilityDeleted", uncommittedEvents[0].EventType())

	eventData := uncommittedEvents[0].EventData()
	assert.Equal(t, capability.ID(), eventData["id"])
	assert.NotNil(t, eventData["deletedAt"])
}

func TestCapability_Delete_LoadFromHistory(t *testing.T) {
	capability := createCapability(t, "Customer Engagement", "L1")

	err := capability.Delete()
	require.NoError(t, err)

	allEvents := capability.GetUncommittedChanges()
	require.Len(t, allEvents, 2)

	loadedCapability, err := LoadCapabilityFromHistory(allEvents)
	require.NoError(t, err)
	assert.Equal(t, capability.ID(), loadedCapability.ID())
	assert.Equal(t, capability.Name().Value(), loadedCapability.Name().Value())
}

func TestCapability_Delete_PreservesAggregateState(t *testing.T) {
	capability := createCapability(t, "Finance", "L1")
	capability.MarkChangesAsCommitted()

	originalID := capability.ID()
	originalName := capability.Name().Value()

	err := capability.Delete()
	require.NoError(t, err)

	assert.Equal(t, originalID, capability.ID())
	assert.Equal(t, originalName, capability.Name().Value())
}

func TestCapability_CanBeAssignedToDomain_L1_Succeeds(t *testing.T) {
	capability := createCapability(t, "Customer Engagement", "L1")

	err := capability.CanBeAssignedToDomain()
	assert.NoError(t, err)
}

func TestCapability_CanBeAssignedToDomain_NonL1_Fails(t *testing.T) {
	for _, level := range []string{"L2", "L3", "L4"} {
		t.Run(level, func(t *testing.T) {
			capability := createCapability(t, "Test Capability", level)

			err := capability.CanBeAssignedToDomain()
			assert.Error(t, err)
			assert.Equal(t, ErrOnlyL1CanBeAssignedToDomain, err)
		})
	}
}

func TestCapability_AddExpert_RaisesExpertAddedEvent(t *testing.T) {
	capability := createCapability(t, "Customer Management", "L1")
	capability.MarkChangesAsCommitted()

	expert := valueobjects.MustNewExpert("Alice Smith", "Product Owner", "alice@example.com", capability.CreatedAt())

	err := capability.AddExpert(expert)
	require.NoError(t, err)

	assertSingleExpertEvent(t, capability, "CapabilityExpertAdded")
}

func TestCapability_RemoveExpert_RaisesExpertRemovedEvent(t *testing.T) {
	capability := createCapability(t, "Customer Management", "L1")
	expert := valueobjects.MustNewExpert("Alice Smith", "Product Owner", "alice@example.com", capability.CreatedAt())
	_ = capability.AddExpert(expert)
	capability.MarkChangesAsCommitted()

	err := capability.RemoveExpert(expert)
	require.NoError(t, err)

	assertSingleExpertEvent(t, capability, "CapabilityExpertRemoved")
}

func assertSingleExpertEvent(t *testing.T, capability *Capability, expectedEventType string) {
	t.Helper()

	uncommittedEvents := capability.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)
	assert.Equal(t, expectedEventType, uncommittedEvents[0].EventType())

	eventData := uncommittedEvents[0].EventData()
	assert.Equal(t, "Alice Smith", eventData["expertName"])
	assert.Equal(t, "Product Owner", eventData["expertRole"])
	assert.Equal(t, "alice@example.com", eventData["contactInfo"])
}

func TestCapability_RemoveExpert_ExpertNoLongerInList(t *testing.T) {
	capability := createCapability(t, "Customer Management", "L1")
	expert1 := valueobjects.MustNewExpert("Alice Smith", "Product Owner", "alice@example.com", capability.CreatedAt())
	expert2 := valueobjects.MustNewExpert("Bob Jones", "Domain Expert", "bob@example.com", capability.CreatedAt())
	_ = capability.AddExpert(expert1)
	_ = capability.AddExpert(expert2)
	capability.MarkChangesAsCommitted()

	err := capability.RemoveExpert(expert1)
	require.NoError(t, err)

	experts := capability.Experts()
	assert.Len(t, experts, 1)
	assert.Equal(t, "Bob Jones", experts[0].Name())
}

func TestCapability_RemoveExpert_LoadFromHistory(t *testing.T) {
	capability := createCapability(t, "Customer Management", "L1")
	expert := valueobjects.MustNewExpert("Alice Smith", "Product Owner", "alice@example.com", capability.CreatedAt())
	_ = capability.AddExpert(expert)
	_ = capability.RemoveExpert(expert)

	allEvents := capability.GetUncommittedChanges()
	require.Len(t, allEvents, 3)

	loadedCapability, err := LoadCapabilityFromHistory(allEvents)
	require.NoError(t, err)

	assert.Equal(t, capability.ID(), loadedCapability.ID())
	assert.Empty(t, loadedCapability.Experts())
}

func TestCapability_AddExpert_CustomRole_SavesRole(t *testing.T) {
	capability := createCapability(t, "Customer Management", "L1")
	capability.MarkChangesAsCommitted()

	expert := valueobjects.MustNewExpert("Jane Doe", "Security Champion", "jane@example.com", capability.CreatedAt())

	err := capability.AddExpert(expert)
	require.NoError(t, err)

	experts := capability.Experts()
	require.Len(t, experts, 1)
	assert.Equal(t, "Security Champion", experts[0].Role())
}

func TestCapability_MultipleExperts_AllPersist(t *testing.T) {
	capability := createCapability(t, "Customer Management", "L1")
	expert1 := valueobjects.MustNewExpert("Alice Smith", "Product Owner", "alice@example.com", capability.CreatedAt())
	expert2 := valueobjects.MustNewExpert("Bob Jones", "Domain Expert", "bob@example.com", capability.CreatedAt())
	_ = capability.AddExpert(expert1)
	_ = capability.AddExpert(expert2)

	allEvents := capability.GetUncommittedChanges()
	loadedCapability, err := LoadCapabilityFromHistory(allEvents)
	require.NoError(t, err)

	experts := loadedCapability.Experts()
	assert.Len(t, experts, 2)

	names := []string{experts[0].Name(), experts[1].Name()}
	assert.Contains(t, names, "Alice Smith")
	assert.Contains(t, names, "Bob Jones")
}
