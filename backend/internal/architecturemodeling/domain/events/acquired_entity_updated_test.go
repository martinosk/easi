package events

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAcquiredEntityUpdated_EventData_ClearedAcquisitionDateIsExplicitNull(t *testing.T) {
	event := NewAcquiredEntityUpdated("entity-1", "Acme", nil, "In Progress", "notes")

	data := event.EventData()

	value, ok := data["acquisitionDate"]
	require.True(t, ok, "acquisitionDate key must be present even when cleared, so a JSONB merge overwrites the stale value")
	assert.Nil(t, value)
}

func TestAcquiredEntityUpdated_EventData_SetAcquisitionDateIsCarried(t *testing.T) {
	date := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	event := NewAcquiredEntityUpdated("entity-1", "Acme", &date, "In Progress", "notes")

	data := event.EventData()

	assert.Equal(t, date, data["acquisitionDate"])
}
