package events

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCapabilityRealizationsInherited_EventData(t *testing.T) {
	linkedAt := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	realizations := []InheritedRealization{
		{
			CapabilityID:         "cap-1",
			ComponentID:          "comp-1",
			ComponentName:        "Component",
			RealizationLevel:     "Full",
			Notes:                "",
			Origin:               "Inherited",
			SourceRealizationID:  "real-1",
			SourceCapabilityID:   "cap-source",
			SourceCapabilityName: "Source",
			LinkedAt:             linkedAt,
		},
	}

	event := NewCapabilityRealizationsInherited("cap-root", realizations)

	assert.Equal(t, "CapabilityRealizationsInherited", event.EventType())
	eventData := event.EventData()

	assert.Equal(t, "cap-root", eventData["capabilityId"])
	stored, ok := eventData["inheritedRealizations"].([]InheritedRealization)
	require.True(t, ok)
	require.Len(t, stored, 1)
	assert.Equal(t, realizations[0], stored[0])
}
