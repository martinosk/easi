package events

import (
	"testing"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"

	"github.com/stretchr/testify/assert"
)

func TestNewJourneyProgressUpdated_CarriesActorAndTimestamp_Rule13(t *testing.T) {
	evt := NewJourneyProgressUpdated(JourneyProgressUpdatedFields{ID: "journey-1", Progress: 60, UpdatedBy: "architect@example.com"})

	assert.Equal(t, "journey-1", evt.ID)
	assert.Equal(t, "journey-1", evt.AggregateID())
	assert.Equal(t, pl.JourneyProgressUpdated, evt.EventType())
	assert.Equal(t, 60, evt.Progress)
	assert.Equal(t, "architect@example.com", evt.UpdatedBy)
	assert.False(t, evt.OccurredOn.IsZero())

	data := evt.EventData()
	assert.Equal(t, 60, data["progress"])
}
