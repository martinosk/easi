package repositories

import (
	"encoding/json"
	"testing"
	"time"

	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/entities"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplicationComponentDeserializers_RoundTrip(t *testing.T) {
	name, _ := valueobjects.NewComponentName("Order Service")
	description := valueobjects.MustNewDescription("Handles order processing")

	original, err := aggregates.NewApplicationComponent(name, description)
	require.NoError(t, err)

	events := original.GetUncommittedChanges()
	require.Len(t, events, 1, "Expected 1 event: Created")

	storedEvents := simulateComponentEventStoreRoundTrip(t, events)
	deserializedEvents, err := componentEventDeserializers.Deserialize(storedEvents)
	require.NoError(t, err)

	require.Len(t, deserializedEvents, 1, "All events should be deserialized")

	loaded, err := aggregates.LoadApplicationComponentFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.Equal(t, original.ID(), loaded.ID())
	assert.Equal(t, original.Name().Value(), loaded.Name().Value())
	assert.Equal(t, original.Description().Value(), loaded.Description().Value())
}

func TestApplicationComponentDeserializers_RoundTripWithUpdate(t *testing.T) {
	name, _ := valueobjects.NewComponentName("Payment Service")
	description := valueobjects.MustNewDescription("Handles payments")

	original, err := aggregates.NewApplicationComponent(name, description)
	require.NoError(t, err)

	newName, _ := valueobjects.NewComponentName("Payment Gateway")
	newDescription := valueobjects.MustNewDescription("Updated payment description")
	_ = original.Update(newName, newDescription)

	events := original.GetUncommittedChanges()
	require.Len(t, events, 2, "Expected 2 events: Created, Updated")

	storedEvents := simulateComponentEventStoreRoundTrip(t, events)
	deserializedEvents, err := componentEventDeserializers.Deserialize(storedEvents)
	require.NoError(t, err)

	require.Len(t, deserializedEvents, 2, "All events should be deserialized")

	loaded, err := aggregates.LoadApplicationComponentFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.Equal(t, newName.Value(), loaded.Name().Value())
	assert.Equal(t, newDescription.Value(), loaded.Description().Value())
}

func TestApplicationComponentDeserializers_AllEventsCanBeDeserialized(t *testing.T) {
	name, _ := valueobjects.NewComponentName("Test Component")
	description := valueobjects.MustNewDescription("Test description")

	component, err := aggregates.NewApplicationComponent(name, description)
	require.NoError(t, err)

	newName, _ := valueobjects.NewComponentName("Updated Name")
	newDescription := valueobjects.MustNewDescription("Updated description")
	_ = component.Update(newName, newDescription)

	expert, _ := entities.NewExpert("Alice Smith", "Product Owner", "alice@example.com")
	_ = component.AddExpert(expert)

	_ = component.RemoveExpert("Alice Smith")

	_ = component.Delete()

	events := component.GetUncommittedChanges()
	require.Len(t, events, 5, "Expected 5 events: Created, Updated, ExpertAdded, ExpertRemoved, Deleted")

	storedEvents := simulateComponentEventStoreRoundTrip(t, events)
	deserializedEvents, err := componentEventDeserializers.Deserialize(storedEvents)
	require.NoError(t, err)

	require.Len(t, deserializedEvents, len(events),
		"All events should be deserialized - missing deserializer for one or more event types")

	for i, originalEvent := range events {
		assert.Equal(t, originalEvent.EventType(), deserializedEvents[i].EventType(),
			"Event type mismatch at index %d", i)
	}
}

type componentStoredEventWrapper struct {
	eventType string
	eventData map[string]interface{}
}

func (e *componentStoredEventWrapper) EventType() string                 { return e.eventType }
func (e *componentStoredEventWrapper) EventData() map[string]interface{} { return e.eventData }
func (e *componentStoredEventWrapper) AggregateID() string               { return "" }
func (e *componentStoredEventWrapper) OccurredAt() time.Time             { return time.Time{} }

func simulateComponentEventStoreRoundTrip(t *testing.T, events []domain.DomainEvent) []domain.DomainEvent {
	t.Helper()

	result := make([]domain.DomainEvent, len(events))

	for i, event := range events {
		jsonBytes, err := json.Marshal(event.EventData())
		require.NoError(t, err, "Failed to serialize event: %s", event.EventType())

		var data map[string]interface{}
		err = json.Unmarshal(jsonBytes, &data)
		require.NoError(t, err, "Failed to unmarshal JSON for event: %s", event.EventType())

		result[i] = &componentStoredEventWrapper{
			eventType: event.EventType(),
			eventData: data,
		}
	}

	return result
}
