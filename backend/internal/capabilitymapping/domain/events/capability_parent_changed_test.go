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
	event := NewCapabilityParentChanged("cap-123", "parent-old", "", "L2", "L1")

	assert.Equal(t, "parent-old", event.OldParentID)
	assert.Empty(t, event.NewParentID)
}

func TestNewCapabilityParentChanged_FromRootToChild(t *testing.T) {
	event := NewCapabilityParentChanged("cap-123", "", "parent-new", "L1", "L2")

	assert.Empty(t, event.OldParentID)
	assert.Equal(t, "parent-new", event.NewParentID)
}

func TestCapabilityParentChanged_EventData(t *testing.T) {
	event := NewCapabilityParentChanged("cap-123", "parent-old", "parent-new", "L2", "L3")
	data := event.EventData()

	assert.Equal(t, "cap-123", data["capabilityId"])
	assert.Equal(t, "parent-old", data["oldParentId"])
	assert.Equal(t, "parent-new", data["newParentId"])
	assert.Equal(t, "L2", data["oldLevel"])
	assert.Equal(t, "L3", data["newLevel"])
	assert.NotNil(t, data["timestamp"])
}
