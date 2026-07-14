package events

import (
	"testing"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJourneyPlanned_CarriesActorAndTimestamp_Rule13(t *testing.T) {
	evt := NewJourneyPlanned(JourneyPlannedFields{
		ID:               "journey-1",
		CapabilityID:     "cap-1",
		Kind:             "migration",
		FromComponentIDs: []string{"seabook"},
		ToComponentID:    "phoenix",
		Note:             "route by route",
		PlannedBy:        "architect@example.com",
	})

	assert.Equal(t, "journey-1", evt.ID)
	assert.Equal(t, "journey-1", evt.AggregateID())
	assert.Equal(t, pl.JourneyPlanned, evt.EventType())
	assert.Equal(t, "cap-1", evt.CapabilityID)
	assert.Equal(t, "migration", evt.Kind)
	assert.Equal(t, []string{"seabook"}, evt.FromComponentIDs)
	assert.Equal(t, "phoenix", evt.ToComponentID)
	assert.Equal(t, "route by route", evt.Note)
	assert.Equal(t, "architect@example.com", evt.PlannedBy)
	assert.False(t, evt.OccurredOn.IsZero())
	assert.Nil(t, evt.TargetPeriod)
}

func TestNewJourneyPlanned_MoveFields(t *testing.T) {
	evt := NewJourneyPlanned(JourneyPlannedFields{
		ID:             "journey-2",
		CapabilityID:   "cap-2",
		Kind:           "move",
		ToComponentID:  "sap-s4",
		TargetPeriod:   &TargetPeriodData{Year: 2027, Quarter: 2},
		TargetDomainID: "domain-1",
		TargetParentID: "parent-1",
		ResultingName:  "Freight invoicing",
		PlannedBy:      "architect@example.com",
	})

	require.NotNil(t, evt.TargetPeriod)
	assert.Equal(t, 2027, evt.TargetPeriod.Year)
	assert.Equal(t, 2, evt.TargetPeriod.Quarter)
	assert.Equal(t, "domain-1", evt.TargetDomainID)
	assert.Equal(t, "parent-1", evt.TargetParentID)
	assert.Equal(t, "Freight invoicing", evt.ResultingName)
}

func TestJourneyPlanned_EventData_NestsTargetPeriod(t *testing.T) {
	evt := NewJourneyPlanned(JourneyPlannedFields{
		ID:            "journey-3",
		CapabilityID:  "cap-3",
		Kind:          "carve-out",
		ToComponentID: "pricing-engine",
		TargetPeriod:  &TargetPeriodData{Year: 2026, Quarter: 4},
		PlannedBy:     "architect@example.com",
	})

	data := evt.EventData()

	require.Equal(t, "journey-3", data["id"])
	period, ok := data["targetPeriod"].(map[string]interface{})
	require.True(t, ok, "targetPeriod must serialize as a nested object")
	assert.Equal(t, 2026, period["year"])
	assert.Equal(t, 4, period["quarter"])
}

func TestJourneyPlanned_EventData_NilTargetPeriod(t *testing.T) {
	evt := NewJourneyPlanned(JourneyPlannedFields{
		ID:            "journey-4",
		CapabilityID:  "cap-4",
		Kind:          "migration",
		ToComponentID: "phoenix",
		PlannedBy:     "architect@example.com",
	})

	data := evt.EventData()

	assert.Nil(t, data["targetPeriod"])
}
