package events

import (
	"testing"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"

	"github.com/stretchr/testify/assert"
)

func TestNewJourneyStarted_CarriesActorAndTimestamp_Rule13(t *testing.T) {
	evt := NewJourneyStarted(JourneyStartedFields{ID: "journey-1", StartedBy: "architect@example.com"})

	assert.Equal(t, "journey-1", evt.ID)
	assert.Equal(t, "journey-1", evt.AggregateID())
	assert.Equal(t, pl.JourneyStarted, evt.EventType())
	assert.Equal(t, "architect@example.com", evt.StartedBy)
	assert.False(t, evt.OccurredOn.IsZero())

	data := evt.EventData()
	assert.Equal(t, "journey-1", data["id"])
	assert.Equal(t, "architect@example.com", data["startedBy"])
}
