package events

import (
	"testing"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJourneyDetailsUpdated_CarriesActorAndTimestamp_Rule13(t *testing.T) {
	evt := NewJourneyDetailsUpdated(JourneyDetailsUpdatedFields{
		ID:            "journey-1",
		Note:          "revised plan",
		TargetPeriod:  &TargetPeriodData{Year: 2027, Quarter: 3},
		ResultingName: "Freight invoicing",
		UpdatedBy:     "architect@example.com",
	})

	assert.Equal(t, "journey-1", evt.ID)
	assert.Equal(t, "journey-1", evt.AggregateID())
	assert.Equal(t, pl.JourneyDetailsUpdated, evt.EventType())
	assert.Equal(t, "revised plan", evt.Note)
	require.NotNil(t, evt.TargetPeriod)
	assert.Equal(t, 2027, evt.TargetPeriod.Year)
	assert.Equal(t, "Freight invoicing", evt.ResultingName)
	assert.Equal(t, "architect@example.com", evt.UpdatedBy)
	assert.False(t, evt.OccurredOn.IsZero())
}

func TestJourneyDetailsUpdated_EventData_NilTargetPeriod(t *testing.T) {
	evt := NewJourneyDetailsUpdated(JourneyDetailsUpdatedFields{ID: "journey-2", Note: "no period", UpdatedBy: "a@example.com"})

	data := evt.EventData()

	assert.Nil(t, data["targetPeriod"])
}
