package events

import (
	"testing"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"

	"github.com/stretchr/testify/assert"
)

func TestNewJourneyMilestonesReordered_CarriesOrderActorAndTimestamp_Rule3(t *testing.T) {
	evt := NewJourneyMilestonesReordered(JourneyMilestonesReorderedFields{
		ID:           "journey-1",
		MilestoneIDs: []string{"m2", "m1", "m3"},
		ReorderedBy:  "architect@example.com",
	})

	assert.Equal(t, "journey-1", evt.ID)
	assert.Equal(t, "journey-1", evt.AggregateID())
	assert.Equal(t, pl.JourneyMilestonesReordered, evt.EventType())
	assert.Equal(t, []string{"m2", "m1", "m3"}, evt.MilestoneIDs)
	assert.Equal(t, "architect@example.com", evt.ReorderedBy)
	assert.False(t, evt.OccurredOn.IsZero())
	assert.Equal(t, []string{"m2", "m1", "m3"}, evt.EventData()["milestoneIds"])
}
