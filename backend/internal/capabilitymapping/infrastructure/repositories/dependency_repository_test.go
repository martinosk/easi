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

func TestDependencyDeserializers_RoundTrip(t *testing.T) {
	sourceID := valueobjects.NewCapabilityID()
	targetID := valueobjects.NewCapabilityID()
	depType, _ := valueobjects.NewDependencyType("uses")
	description := valueobjects.MustNewDescription("Source uses target")

	original, err := aggregates.NewCapabilityDependency(sourceID, targetID, depType, description)
	require.NoError(t, err)

	events := original.GetUncommittedChanges()
	require.Len(t, events, 1, "Expected 1 event: Created")

	storedEvents := simulateDependencyEventStoreRoundTrip(t, events)
	deserializedEvents, err := dependencyEventDeserializers.Deserialize(storedEvents)
	require.NoError(t, err)

	require.Len(t, deserializedEvents, 1, "All events should be deserialized")

	loaded, err := aggregates.LoadCapabilityDependencyFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.Equal(t, original.ID(), loaded.ID())
	assert.Equal(t, original.SourceCapabilityID().Value(), loaded.SourceCapabilityID().Value())
	assert.Equal(t, original.TargetCapabilityID().Value(), loaded.TargetCapabilityID().Value())
	assert.Equal(t, original.DependencyType().Value(), loaded.DependencyType().Value())
	assert.Equal(t, original.Description().Value(), loaded.Description().Value())
}

func TestDependencyDeserializers_AllEventsCanBeDeserialized(t *testing.T) {
	sourceID := valueobjects.NewCapabilityID()
	targetID := valueobjects.NewCapabilityID()
	depType, _ := valueobjects.NewDependencyType("depends-on")
	description := valueobjects.MustNewDescription("Dependency description")

	dependency, err := aggregates.NewCapabilityDependency(sourceID, targetID, depType, description)
	require.NoError(t, err)

	_ = dependency.Delete()

	events := dependency.GetUncommittedChanges()
	require.Len(t, events, 2, "Expected 2 events: Created, Deleted")

	storedEvents := simulateDependencyEventStoreRoundTrip(t, events)
	deserializedEvents, err := dependencyEventDeserializers.Deserialize(storedEvents)
	require.NoError(t, err)

	require.Len(t, deserializedEvents, len(events),
		"All events should be deserialized - missing deserializer for one or more event types")

	for i, originalEvent := range events {
		assert.Equal(t, originalEvent.EventType(), deserializedEvents[i].EventType(),
			"Event type mismatch at index %d", i)
	}
}

type dependencyStoredEventWrapper struct {
	eventType string
	eventData map[string]interface{}
}

func (e *dependencyStoredEventWrapper) EventType() string                 { return e.eventType }
func (e *dependencyStoredEventWrapper) EventData() map[string]interface{} { return e.eventData }
func (e *dependencyStoredEventWrapper) AggregateID() string               { return "" }
func (e *dependencyStoredEventWrapper) OccurredAt() time.Time             { return time.Time{} }

func simulateDependencyEventStoreRoundTrip(t *testing.T, events []domain.DomainEvent) []domain.DomainEvent {
	t.Helper()

	result := make([]domain.DomainEvent, len(events))

	for i, event := range events {
		jsonBytes, err := json.Marshal(event.EventData())
		require.NoError(t, err, "Failed to serialize event: %s", event.EventType())

		var data map[string]interface{}
		err = json.Unmarshal(jsonBytes, &data)
		require.NoError(t, err, "Failed to unmarshal JSON for event: %s", event.EventType())

		result[i] = &dependencyStoredEventWrapper{
			eventType: event.EventType(),
			eventData: data,
		}
	}

	return result
}
