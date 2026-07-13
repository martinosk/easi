package events

import (
	"testing"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"

	"github.com/stretchr/testify/assert"
)

func TestNewJourneySourceApplicationsChanged_CarriesActorAndTimestamp_Rule13(t *testing.T) {
	evt := NewJourneySourceApplicationsChanged(JourneySourceApplicationsChangedFields{
		ID:               "journey-1",
		FromComponentIDs: []string{"seabook", "legacy-booking"},
		ChangedBy:        "architect@example.com",
	})

	assert.Equal(t, "journey-1", evt.ID)
	assert.Equal(t, "journey-1", evt.AggregateID())
	assert.Equal(t, pl.JourneySourceApplicationsChanged, evt.EventType())
	assert.Equal(t, []string{"seabook", "legacy-booking"}, evt.FromComponentIDs)
	assert.Equal(t, "architect@example.com", evt.ChangedBy)
	assert.False(t, evt.OccurredOn.IsZero())
}
