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

func TestBusinessDomainDeserializers_RoundTrip(t *testing.T) {
	name, _ := valueobjects.NewDomainName("Sales")
	description := valueobjects.MustNewDescription("Sales domain operations")

	original, err := aggregates.NewBusinessDomain(name, description)
	require.NoError(t, err)

	events := original.GetUncommittedChanges()
	require.Len(t, events, 1, "Expected 1 event: Created")

	storedEvents := simulateBusinessDomainEventStoreRoundTrip(t, events)
	deserializedEvents, err := businessDomainEventDeserializers.Deserialize(storedEvents)
	require.NoError(t, err)

	require.Len(t, deserializedEvents, 1, "All events should be deserialized")

	loaded, err := aggregates.LoadBusinessDomainFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.Equal(t, original.ID(), loaded.ID())
	assert.Equal(t, original.Name().Value(), loaded.Name().Value())
	assert.Equal(t, original.Description().Value(), loaded.Description().Value())
}

func TestBusinessDomainDeserializers_RoundTripWithUpdate(t *testing.T) {
	name, _ := valueobjects.NewDomainName("Marketing")
	description := valueobjects.MustNewDescription("Marketing operations")

	original, err := aggregates.NewBusinessDomain(name, description)
	require.NoError(t, err)

	newName, _ := valueobjects.NewDomainName("Digital Marketing")
	newDescription := valueobjects.MustNewDescription("Updated marketing description")
	_ = original.Update(newName, newDescription)

	events := original.GetUncommittedChanges()
	require.Len(t, events, 2, "Expected 2 events: Created, Updated")

	storedEvents := simulateBusinessDomainEventStoreRoundTrip(t, events)
	deserializedEvents, err := businessDomainEventDeserializers.Deserialize(storedEvents)
	require.NoError(t, err)

	require.Len(t, deserializedEvents, 2, "All events should be deserialized")

	loaded, err := aggregates.LoadBusinessDomainFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.Equal(t, newName.Value(), loaded.Name().Value())
	assert.Equal(t, newDescription.Value(), loaded.Description().Value())
}

func TestBusinessDomainDeserializers_AllEventsCanBeDeserialized(t *testing.T) {
	name, _ := valueobjects.NewDomainName("Finance")
	description := valueobjects.MustNewDescription("Finance operations")

	businessDomain, err := aggregates.NewBusinessDomain(name, description)
	require.NoError(t, err)

	newName, _ := valueobjects.NewDomainName("Financial Services")
	newDescription := valueobjects.MustNewDescription("Updated finance description")
	_ = businessDomain.Update(newName, newDescription)

	_ = businessDomain.Delete()

	events := businessDomain.GetUncommittedChanges()
	require.Len(t, events, 3, "Expected 3 events: Created, Updated, Deleted")

	storedEvents := simulateBusinessDomainEventStoreRoundTrip(t, events)
	deserializedEvents, err := businessDomainEventDeserializers.Deserialize(storedEvents)
	require.NoError(t, err)

	require.Len(t, deserializedEvents, len(events),
		"All events should be deserialized - missing deserializer for one or more event types")

	for i, originalEvent := range events {
		assert.Equal(t, originalEvent.EventType(), deserializedEvents[i].EventType(),
			"Event type mismatch at index %d", i)
	}
}

type businessDomainStoredEventWrapper struct {
	eventType string
	eventData map[string]interface{}
}

func (e *businessDomainStoredEventWrapper) EventType() string                 { return e.eventType }
func (e *businessDomainStoredEventWrapper) EventData() map[string]interface{} { return e.eventData }
func (e *businessDomainStoredEventWrapper) AggregateID() string               { return "" }
func (e *businessDomainStoredEventWrapper) OccurredAt() time.Time             { return time.Time{} }

func simulateBusinessDomainEventStoreRoundTrip(t *testing.T, events []domain.DomainEvent) []domain.DomainEvent {
	t.Helper()

	result := make([]domain.DomainEvent, len(events))

	for i, event := range events {
		jsonBytes, err := json.Marshal(event.EventData())
		require.NoError(t, err, "Failed to serialize event: %s", event.EventType())

		var data map[string]interface{}
		err = json.Unmarshal(jsonBytes, &data)
		require.NoError(t, err, "Failed to unmarshal JSON for event: %s", event.EventType())

		result[i] = &businessDomainStoredEventWrapper{
			eventType: event.EventType(),
			eventData: data,
		}
	}

	return result
}
