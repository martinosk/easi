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

func TestRealizationDeserializers_RoundTrip(t *testing.T) {
	capabilityID := valueobjects.NewCapabilityID()
	componentID, _ := valueobjects.NewComponentIDFromString("550e8400-e29b-41d4-a716-446655440000")
	realizationLevel, _ := valueobjects.NewRealizationLevel("primary")
	notes := valueobjects.MustNewDescription("Main implementation")

	original, err := aggregates.NewCapabilityRealization(capabilityID, componentID, realizationLevel, notes)
	require.NoError(t, err)

	events := original.GetUncommittedChanges()
	require.Len(t, events, 1, "Expected 1 event: SystemLinkedToCapability")

	storedEvents := simulateRealizationEventStoreRoundTrip(t, events)
	deserializedEvents := realizationEventDeserializers.Deserialize(storedEvents)

	require.Len(t, deserializedEvents, 1, "All events should be deserialized")

	loaded, err := aggregates.LoadCapabilityRealizationFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.Equal(t, original.ID(), loaded.ID())
	assert.Equal(t, original.CapabilityID().Value(), loaded.CapabilityID().Value())
	assert.Equal(t, original.ComponentID().Value(), loaded.ComponentID().Value())
	assert.Equal(t, original.RealizationLevel().Value(), loaded.RealizationLevel().Value())
	assert.Equal(t, original.Notes().Value(), loaded.Notes().Value())
}

func TestRealizationDeserializers_RoundTripWithUpdate(t *testing.T) {
	capabilityID := valueobjects.NewCapabilityID()
	componentID, _ := valueobjects.NewComponentIDFromString("550e8400-e29b-41d4-a716-446655440001")
	realizationLevel, _ := valueobjects.NewRealizationLevel("supporting")
	notes := valueobjects.MustNewDescription("Initial notes")

	original, err := aggregates.NewCapabilityRealization(capabilityID, componentID, realizationLevel, notes)
	require.NoError(t, err)

	newLevel, _ := valueobjects.NewRealizationLevel("primary")
	newNotes := valueobjects.MustNewDescription("Updated notes")
	_ = original.Update(newLevel, newNotes)

	events := original.GetUncommittedChanges()
	require.Len(t, events, 2, "Expected 2 events: Linked, Updated")

	storedEvents := simulateRealizationEventStoreRoundTrip(t, events)
	deserializedEvents := realizationEventDeserializers.Deserialize(storedEvents)

	require.Len(t, deserializedEvents, 2, "All events should be deserialized")

	loaded, err := aggregates.LoadCapabilityRealizationFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.Equal(t, newLevel.Value(), loaded.RealizationLevel().Value())
	assert.Equal(t, newNotes.Value(), loaded.Notes().Value())
}

func TestRealizationDeserializers_AllEventsCanBeDeserialized(t *testing.T) {
	capabilityID := valueobjects.NewCapabilityID()
	componentID, _ := valueobjects.NewComponentIDFromString("550e8400-e29b-41d4-a716-446655440002")
	realizationLevel, _ := valueobjects.NewRealizationLevel("primary")
	notes := valueobjects.MustNewDescription("Test notes")

	realization, err := aggregates.NewCapabilityRealization(capabilityID, componentID, realizationLevel, notes)
	require.NoError(t, err)

	newLevel, _ := valueobjects.NewRealizationLevel("supporting")
	newNotes := valueobjects.MustNewDescription("Updated notes")
	_ = realization.Update(newLevel, newNotes)

	_ = realization.Delete()

	events := realization.GetUncommittedChanges()
	require.Len(t, events, 3, "Expected 3 events: Linked, Updated, Deleted")

	storedEvents := simulateRealizationEventStoreRoundTrip(t, events)
	deserializedEvents := realizationEventDeserializers.Deserialize(storedEvents)

	require.Len(t, deserializedEvents, len(events),
		"All events should be deserialized - missing deserializer for one or more event types")

	for i, originalEvent := range events {
		assert.Equal(t, originalEvent.EventType(), deserializedEvents[i].EventType(),
			"Event type mismatch at index %d", i)
	}
}

type realizationStoredEventWrapper struct {
	eventType string
	eventData map[string]interface{}
}

func (e *realizationStoredEventWrapper) EventType() string                 { return e.eventType }
func (e *realizationStoredEventWrapper) EventData() map[string]interface{} { return e.eventData }
func (e *realizationStoredEventWrapper) AggregateID() string               { return "" }
func (e *realizationStoredEventWrapper) OccurredAt() time.Time             { return time.Time{} }

func simulateRealizationEventStoreRoundTrip(t *testing.T, events []domain.DomainEvent) []domain.DomainEvent {
	t.Helper()

	result := make([]domain.DomainEvent, len(events))

	for i, event := range events {
		jsonBytes, err := json.Marshal(event.EventData())
		require.NoError(t, err, "Failed to serialize event: %s", event.EventType())

		var data map[string]interface{}
		err = json.Unmarshal(jsonBytes, &data)
		require.NoError(t, err, "Failed to unmarshal JSON for event: %s", event.EventType())

		result[i] = &realizationStoredEventWrapper{
			eventType: event.EventType(),
			eventData: data,
		}
	}

	return result
}
