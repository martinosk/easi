package events

import (
	"testing"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJourneyMilestoneAdded_CarriesActorAndTimestamp_Rule13(t *testing.T) {
	evt := NewJourneyMilestoneAdded(JourneyMilestoneFields{
		ID:           "journey-1",
		MilestoneID:  "milestone-1",
		Label:        "Route cutover",
		TargetPeriod: &TargetPeriodData{Year: 2026, Quarter: 4},
		Status:       "planned",
		Actor:        "architect@example.com",
	})

	assert.Equal(t, "journey-1", evt.ID)
	assert.Equal(t, "journey-1", evt.AggregateID())
	assert.Equal(t, pl.JourneyMilestoneAdded, evt.EventType())
	assert.Equal(t, "milestone-1", evt.MilestoneID)
	assert.Equal(t, "Route cutover", evt.Label)
	require.NotNil(t, evt.TargetPeriod)
	assert.Equal(t, "planned", evt.Status)
	assert.Equal(t, "architect@example.com", evt.AddedBy)
	assert.False(t, evt.OccurredOn.IsZero())
}
