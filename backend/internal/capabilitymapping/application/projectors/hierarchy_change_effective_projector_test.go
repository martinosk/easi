package projectors

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHierarchyChangeEffectiveProjector_CapabilityParentChanged_ChecksNewParentImportances(t *testing.T) {
	projector := &HierarchyChangeEffectiveProjector{}

	capabilityParentChangedEvent := map[string]interface{}{
		"capabilityId": "child-cap-id",
		"oldParentId":  "",
		"newParentId":  "parent-cap-id",
		"oldLevel":     "L1",
		"newLevel":     "L2",
	}
	eventData, err := json.Marshal(capabilityParentChangedEvent)
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "CapabilityParentChanged", eventData)

	assert.Error(t, err, "CapabilityParentChanged should check both child's AND new parent's effective importances")
}

func TestHierarchyChangeEffectiveProjector_CapabilityParentChanged_NoNewParent_OnlyChecksChild(t *testing.T) {
	projector := &HierarchyChangeEffectiveProjector{}

	capabilityParentChangedEvent := map[string]interface{}{
		"capabilityId": "child-cap-id",
		"oldParentId":  "old-parent-id",
		"newParentId":  "",
		"oldLevel":     "L2",
		"newLevel":     "L1",
	}
	eventData, err := json.Marshal(capabilityParentChangedEvent)
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "CapabilityParentChanged", eventData)

	assert.Error(t, err, "CapabilityParentChanged without new parent should still check child's effective importances")
}

func TestHierarchyChangeEffectiveProjector_CapabilityDeleted_MissingIdentifier_ReturnsError(t *testing.T) {
	projector := &HierarchyChangeEffectiveProjector{}

	eventData, err := json.Marshal(map[string]interface{}{"deletedAt": "2026-01-01T00:00:00Z"})
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "CapabilityDeleted", eventData)
	assert.Error(t, err)
}
