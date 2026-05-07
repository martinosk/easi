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

type realizationFixture struct {
	componentUUID    string
	realizationLevel string
	componentName    string
	notes            string
}

func newRealizationFromFixture(t *testing.T, f realizationFixture) *aggregates.CapabilityRealization {
	t.Helper()

	componentID, err := valueobjects.NewComponentIDFromString(f.componentUUID)
	require.NoError(t, err)
	level, err := valueobjects.NewRealizationLevel(f.realizationLevel)
	require.NoError(t, err)

	realization, err := aggregates.NewCapabilityRealization(
		valueobjects.NewCapabilityID(),
		componentID,
		f.componentName,
		level,
		valueobjects.MustNewDescription(f.notes),
	)
	require.NoError(t, err)

	return realization
}

func roundTripDeserializeRealization(t *testing.T, events []domain.DomainEvent) []domain.DomainEvent {
	t.Helper()

	storedEvents := simulateRealizationEventStoreRoundTrip(t, events)
	deserialized, err := realizationEventDeserializers.Deserialize(storedEvents)
	require.NoError(t, err)

	return deserialized
}

func TestRealizationDeserializers_RoundTrip(t *testing.T) {
	original := newRealizationFromFixture(t, realizationFixture{
		componentUUID:    "550e8400-e29b-41d4-a716-446655440000",
		realizationLevel: "Full",
		componentName:    "Test Component",
		notes:            "Main implementation",
	})

	events := original.GetUncommittedChanges()
	require.Len(t, events, 1, "Expected 1 event: SystemLinkedToCapability")

	deserializedEvents := roundTripDeserializeRealization(t, events)
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
	original := newRealizationFromFixture(t, realizationFixture{
		componentUUID:    "550e8400-e29b-41d4-a716-446655440001",
		realizationLevel: "Partial",
		componentName:    "Test Component",
		notes:            "Initial notes",
	})

	newLevel, _ := valueobjects.NewRealizationLevel("Full")
	newNotes := valueobjects.MustNewDescription("Updated notes")
	_ = original.Update(newLevel, newNotes)

	events := original.GetUncommittedChanges()
	require.Len(t, events, 2, "Expected 2 events: Linked, Updated")

	deserializedEvents := roundTripDeserializeRealization(t, events)
	require.Len(t, deserializedEvents, 2, "All events should be deserialized")

	loaded, err := aggregates.LoadCapabilityRealizationFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.Equal(t, newLevel.Value(), loaded.RealizationLevel().Value())
	assert.Equal(t, newNotes.Value(), loaded.Notes().Value())
}

func TestRealizationDeserializers_AllEventsCanBeDeserialized(t *testing.T) {
	realization := newRealizationFromFixture(t, realizationFixture{
		componentUUID:    "550e8400-e29b-41d4-a716-446655440002",
		realizationLevel: "Full",
		componentName:    "Test Component",
		notes:            "Test notes",
	})

	newLevel, _ := valueobjects.NewRealizationLevel("Partial")
	_ = realization.Update(newLevel, valueobjects.MustNewDescription("Updated notes"))
	_ = realization.Delete()

	events := realization.GetUncommittedChanges()
	require.Len(t, events, 3, "Expected 3 events: Linked, Updated, Deleted")

	deserializedEvents := roundTripDeserializeRealization(t, events)
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
