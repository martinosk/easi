package events

import (
	"testing"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"

	"github.com/stretchr/testify/assert"
)

func TestNewJourneyMilestoneRemoved_CarriesActorAndTimestamp_Rule13(t *testing.T) {
	evt := NewJourneyMilestoneRemoved(JourneyMilestoneRemovedFields{
		ID:          "journey-1",
		MilestoneID: "milestone-1",
		RemovedBy:   "architect@example.com",
	})

	assert.Equal(t, "journey-1", evt.ID)
	assert.Equal(t, "journey-1", evt.AggregateID())
	assert.Equal(t, pl.JourneyMilestoneRemoved, evt.EventType())
	assert.Equal(t, "milestone-1", evt.MilestoneID)
	assert.Equal(t, "architect@example.com", evt.RemovedBy)
	assert.False(t, evt.OccurredOn.IsZero())
}
