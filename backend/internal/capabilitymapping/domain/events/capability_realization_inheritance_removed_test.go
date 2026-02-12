package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCapabilityRealizationsUninherited_EventData(t *testing.T) {
	removals := []RealizationInheritanceRemoval{
		{SourceRealizationID: "real-1", CapabilityIDs: []string{"cap-1", "cap-2"}},
	}

	event := NewCapabilityRealizationsUninherited("cap-root", removals)

	assert.Equal(t, "CapabilityRealizationsUninherited", event.EventType())
	eventData := event.EventData()

	assert.Equal(t, "cap-root", eventData["capabilityId"])
	stored, ok := eventData["removals"].([]RealizationInheritanceRemoval)
	require.True(t, ok)
	require.Len(t, stored, 1)
	assert.Equal(t, removals[0], stored[0])
}
