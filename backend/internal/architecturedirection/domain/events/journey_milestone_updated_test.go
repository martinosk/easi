package events

import (
	"testing"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"

	"github.com/stretchr/testify/assert"
)

func TestNewJourneyMilestoneUpdated_CarriesActorAndTimestamp_Rule13(t *testing.T) {
	evt := NewJourneyMilestoneUpdated(JourneyMilestoneFields{
		ID:          "journey-1",
		MilestoneID: "milestone-1",
		Label:       "Route cutover",
		Status:      "done",
		Actor:       "architect@example.com",
	})

	assert.Equal(t, "journey-1", evt.ID)
	assert.Equal(t, "journey-1", evt.AggregateID())
	assert.Equal(t, pl.JourneyMilestoneUpdated, evt.EventType())
	assert.Equal(t, "milestone-1", evt.MilestoneID)
	assert.Equal(t, "done", evt.Status)
	assert.Equal(t, "architect@example.com", evt.UpdatedBy)
	assert.False(t, evt.OccurredOn.IsZero())
}
