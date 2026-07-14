package events

import (
	"testing"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"

	"github.com/stretchr/testify/assert"
)

func TestNewJourneyCompleted_CarriesActorAndTimestamp_Rule13(t *testing.T) {
	evt := NewJourneyCompleted(JourneyCompletedFields{ID: "journey-1", CompletedBy: "architect@example.com"})

	assert.Equal(t, "journey-1", evt.ID)
	assert.Equal(t, "journey-1", evt.AggregateID())
	assert.Equal(t, pl.JourneyCompleted, evt.EventType())
	assert.Equal(t, "architect@example.com", evt.CompletedBy)
	assert.False(t, evt.OccurredOn.IsZero())

	data := evt.EventData()
	assert.Equal(t, "journey-1", data["id"])
	assert.Equal(t, "architect@example.com", data["completedBy"])
}
