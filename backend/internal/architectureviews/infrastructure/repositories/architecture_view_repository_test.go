package repositories

import (
	"encoding/json"
	"testing"
	"time"

	"easi/backend/internal/architectureviews/domain/aggregates"
	"easi/backend/internal/architectureviews/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testOwner() valueobjects.ViewOwner {
	owner, _ := valueobjects.NewViewOwner("test-user-id", "test@example.com")
	return owner
}

func newArchitectureView(t *testing.T, nameStr, description string, isDefault bool) *aggregates.ArchitectureView {
	t.Helper()

	name, err := valueobjects.NewViewName(nameStr)
	require.NoError(t, err)

	view, err := aggregates.NewArchitectureView(name, description, isDefault, testOwner())
	require.NoError(t, err)

	return view
}

func roundTripDeserializeView(t *testing.T, events []domain.DomainEvent) []domain.DomainEvent {
	t.Helper()

	storedEvents := simulateViewEventStoreRoundTrip(t, events)
	deserialized, err := eventDeserializers.Deserialize(storedEvents)
	require.NoError(t, err)

	return deserialized
}

func TestArchitectureViewDeserializers_RoundTrip(t *testing.T) {
	original := newArchitectureView(t, "Main View", "Main architecture view", false)

	events := original.GetUncommittedChanges()
	require.Len(t, events, 1, "Expected 1 event: ViewCreated")

	deserializedEvents := roundTripDeserializeView(t, events)
	require.Len(t, deserializedEvents, 1, "All events should be deserialized")

	loaded, err := aggregates.LoadArchitectureViewFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.Equal(t, original.ID(), loaded.ID())
	assert.Equal(t, original.Name().Value(), loaded.Name().Value())
	assert.Equal(t, original.Description(), loaded.Description())
}

func TestArchitectureViewDeserializers_RoundTripWithDefault(t *testing.T) {
	original := newArchitectureView(t, "Default View", "Default architecture view", true)

	events := original.GetUncommittedChanges()
	require.Len(t, events, 2, "Expected 2 events: ViewCreated, DefaultViewChanged")

	deserializedEvents := roundTripDeserializeView(t, events)
	require.Len(t, deserializedEvents, 2, "All events should be deserialized")

	loaded, err := aggregates.LoadArchitectureViewFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.True(t, loaded.IsDefault())
}

func TestArchitectureViewDeserializers_RoundTripWithComponentAndRename(t *testing.T) {
	original := newArchitectureView(t, "Component View", "View with components", false)

	_ = original.AddComponent("component-1")
	_ = original.AddComponent("component-2")

	newName, _ := valueobjects.NewViewName("Renamed View")
	_ = original.Rename(newName)

	events := original.GetUncommittedChanges()
	require.Len(t, events, 4, "Expected 4 events: ViewCreated, 2x ComponentAdded, ViewRenamed")

	deserializedEvents := roundTripDeserializeView(t, events)
	require.Len(t, deserializedEvents, 4, "All events should be deserialized")

	loaded, err := aggregates.LoadArchitectureViewFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.Equal(t, newName.Value(), loaded.Name().Value())
	assert.True(t, loaded.HasComponent("component-1"))
	assert.True(t, loaded.HasComponent("component-2"))
}

func TestArchitectureViewDeserializers_AllEventsCanBeDeserialized(t *testing.T) {
	view := newArchitectureView(t, "Test View", "Test description", false)

	_ = view.AddComponent("component-1")
	_ = view.RemoveComponent("component-1")

	newName, _ := valueobjects.NewViewName("Renamed Test View")
	_ = view.Rename(newName)

	_ = view.SetAsDefault()
	_ = view.UnsetAsDefault()

	_ = view.Delete()

	events := view.GetUncommittedChanges()
	require.GreaterOrEqual(t, len(events), 6, "Expected at least 6 events")

	deserializedEvents := roundTripDeserializeView(t, events)
	require.Len(t, deserializedEvents, len(events),
		"All events should be deserialized - missing deserializer for one or more event types")

	for i, originalEvent := range events {
		assert.Equal(t, originalEvent.EventType(), deserializedEvents[i].EventType(),
			"Event type mismatch at index %d", i)
	}
}

type viewStoredEventWrapper struct {
	eventType string
	eventData map[string]interface{}
}

func (e *viewStoredEventWrapper) EventType() string                 { return e.eventType }
func (e *viewStoredEventWrapper) EventData() map[string]interface{} { return e.eventData }
func (e *viewStoredEventWrapper) AggregateID() string               { return "" }
func (e *viewStoredEventWrapper) OccurredAt() time.Time             { return time.Time{} }

func simulateViewEventStoreRoundTrip(t *testing.T, events []domain.DomainEvent) []domain.DomainEvent {
	t.Helper()

	result := make([]domain.DomainEvent, len(events))

	for i, event := range events {
		jsonBytes, err := json.Marshal(event.EventData())
		require.NoError(t, err, "Failed to serialize event: %s", event.EventType())

		var data map[string]interface{}
		err = json.Unmarshal(jsonBytes, &data)
		require.NoError(t, err, "Failed to unmarshal JSON for event: %s", event.EventType())

		result[i] = &viewStoredEventWrapper{
			eventType: event.EventType(),
			eventData: data,
		}
	}

	return result
}
