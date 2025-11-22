package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCapabilityParentChanged_WithNewParent(t *testing.T) {
	capabilityID := "cap-123"
	oldParentID := "parent-old"
	newParentID := "parent-new"
	oldLevel := "L2"
	newLevel := "L3"

	event := NewCapabilityParentChanged(capabilityID, oldParentID, newParentID, oldLevel, newLevel)

	assert.Equal(t, capabilityID, event.CapabilityID)
	assert.Equal(t, oldParentID, event.OldParentID)
	assert.Equal(t, newParentID, event.NewParentID)
	assert.Equal(t, oldLevel, event.OldLevel)
	assert.Equal(t, newLevel, event.NewLevel)
	assert.NotZero(t, event.Timestamp)
}

func TestNewCapabilityParentChanged_MovingToRoot(t *testing.T) {
	capabilityID := "cap-123"
	oldParentID := "parent-old"
	newParentID := ""
	oldLevel := "L2"
	newLevel := "L1"

	event := NewCapabilityParentChanged(capabilityID, oldParentID, newParentID, oldLevel, newLevel)

	assert.Equal(t, capabilityID, event.CapabilityID)
	assert.Equal(t, oldParentID, event.OldParentID)
	assert.Empty(t, event.NewParentID)
	assert.Equal(t, oldLevel, event.OldLevel)
	assert.Equal(t, newLevel, event.NewLevel)
}

func TestNewCapabilityParentChanged_FromRootToChild(t *testing.T) {
	capabilityID := "cap-123"
	oldParentID := ""
	newParentID := "parent-new"
	oldLevel := "L1"
	newLevel := "L2"

	event := NewCapabilityParentChanged(capabilityID, oldParentID, newParentID, oldLevel, newLevel)

	assert.Equal(t, capabilityID, event.CapabilityID)
	assert.Empty(t, event.OldParentID)
	assert.Equal(t, newParentID, event.NewParentID)
	assert.Equal(t, oldLevel, event.OldLevel)
	assert.Equal(t, newLevel, event.NewLevel)
}

func TestCapabilityParentChanged_EventType(t *testing.T) {
	event := NewCapabilityParentChanged("cap-123", "old", "new", "L1", "L2")

	assert.Equal(t, "CapabilityParentChanged", event.EventType())
}

func TestCapabilityParentChanged_EventData(t *testing.T) {
	capabilityID := "cap-123"
	oldParentID := "parent-old"
	newParentID := "parent-new"
	oldLevel := "L2"
	newLevel := "L3"

	event := NewCapabilityParentChanged(capabilityID, oldParentID, newParentID, oldLevel, newLevel)
	data := event.EventData()

	assert.Equal(t, capabilityID, data["capabilityId"])
	assert.Equal(t, oldParentID, data["oldParentId"])
	assert.Equal(t, newParentID, data["newParentId"])
	assert.Equal(t, oldLevel, data["oldLevel"])
	assert.Equal(t, newLevel, data["newLevel"])
	assert.NotNil(t, data["timestamp"])
}

func TestCapabilityParentChanged_EventData_ContainsAllFields(t *testing.T) {
	event := NewCapabilityParentChanged("cap-123", "old", "new", "L1", "L2")
	data := event.EventData()

	expectedKeys := []string{"capabilityId", "oldParentId", "newParentId", "oldLevel", "newLevel", "timestamp"}
	for _, key := range expectedKeys {
		_, exists := data[key]
		assert.True(t, exists, "EventData should contain key: %s", key)
	}
	assert.Len(t, data, len(expectedKeys))
}

func TestCapabilityParentChanged_AggregateID(t *testing.T) {
	capabilityID := "cap-123"
	event := NewCapabilityParentChanged(capabilityID, "old", "new", "L1", "L2")

	assert.Equal(t, capabilityID, event.AggregateID())
}

func TestNewCapabilityParentChanged_SameLevelDifferentParent(t *testing.T) {
	capabilityID := "cap-123"
	oldParentID := "parent-old"
	newParentID := "parent-new"
	oldLevel := "L2"
	newLevel := "L2"

	event := NewCapabilityParentChanged(capabilityID, oldParentID, newParentID, oldLevel, newLevel)

	assert.Equal(t, oldLevel, event.OldLevel)
	assert.Equal(t, newLevel, event.NewLevel)
	assert.NotEqual(t, event.OldParentID, event.NewParentID)
}
