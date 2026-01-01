package repositories

import (
	"encoding/json"
	"testing"
	"time"

	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBusinessDomainAssignmentDeserializers_RoundTrip(t *testing.T) {
	businessDomainID := valueobjects.NewBusinessDomainID()
	capabilityID := valueobjects.NewCapabilityID()

	original, err := aggregates.AssignCapabilityToDomain(businessDomainID, capabilityID)
	require.NoError(t, err)

	events := original.GetUncommittedChanges()
	require.Len(t, events, 1, "Expected 1 event: Assigned")

	storedEvents := simulateAssignmentEventStoreRoundTrip(t, events)
	deserializedEvents := assignmentEventDeserializers.Deserialize(storedEvents)

	require.Len(t, deserializedEvents, 1, "All events should be deserialized")

	loaded, err := aggregates.LoadBusinessDomainAssignmentFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.Equal(t, original.ID(), loaded.ID())
	assert.Equal(t, original.BusinessDomainID().Value(), loaded.BusinessDomainID().Value())
	assert.Equal(t, original.CapabilityID().Value(), loaded.CapabilityID().Value())
}

func TestBusinessDomainAssignmentDeserializers_AllEventsCanBeDeserialized(t *testing.T) {
	businessDomainID := valueobjects.NewBusinessDomainID()
	capabilityID := valueobjects.NewCapabilityID()

	assignment, err := aggregates.AssignCapabilityToDomain(businessDomainID, capabilityID)
	require.NoError(t, err)

	_ = assignment.Unassign()

	events := assignment.GetUncommittedChanges()
	require.Len(t, events, 2, "Expected 2 events: Assigned, Unassigned")

	storedEvents := simulateAssignmentEventStoreRoundTrip(t, events)
	deserializedEvents := assignmentEventDeserializers.Deserialize(storedEvents)

	require.Len(t, deserializedEvents, len(events),
		"All events should be deserialized - missing deserializer for one or more event types")

	for i, originalEvent := range events {
		assert.Equal(t, originalEvent.EventType(), deserializedEvents[i].EventType(),
			"Event type mismatch at index %d", i)
	}
}

type assignmentStoredEventWrapper struct {
	eventType string
	eventData map[string]interface{}
}

func (e *assignmentStoredEventWrapper) EventType() string                 { return e.eventType }
func (e *assignmentStoredEventWrapper) EventData() map[string]interface{} { return e.eventData }
func (e *assignmentStoredEventWrapper) AggregateID() string               { return "" }
func (e *assignmentStoredEventWrapper) OccurredAt() time.Time             { return time.Time{} }

func simulateAssignmentEventStoreRoundTrip(t *testing.T, events []domain.DomainEvent) []domain.DomainEvent {
	t.Helper()

	result := make([]domain.DomainEvent, len(events))

	for i, event := range events {
		jsonBytes, err := json.Marshal(event.EventData())
		require.NoError(t, err, "Failed to serialize event: %s", event.EventType())

		var data map[string]interface{}
		err = json.Unmarshal(jsonBytes, &data)
		require.NoError(t, err, "Failed to unmarshal JSON for event: %s", event.EventType())

		result[i] = &assignmentStoredEventWrapper{
			eventType: event.EventType(),
			eventData: data,
		}
	}

	return result
}
