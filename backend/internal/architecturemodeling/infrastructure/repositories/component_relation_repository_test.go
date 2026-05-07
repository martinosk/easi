package repositories

import (
	"encoding/json"
	"testing"
	"time"

	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type relationFixture struct {
	sourceUUID   string
	targetUUID   string
	relationType string
	name         string
	description  string
}

func newRelationFromFixture(t *testing.T, f relationFixture) *aggregates.ComponentRelation {
	t.Helper()

	sourceID, err := valueobjects.NewComponentIDFromString(f.sourceUUID)
	require.NoError(t, err)
	targetID, err := valueobjects.NewComponentIDFromString(f.targetUUID)
	require.NoError(t, err)
	relType, err := valueobjects.NewRelationType(f.relationType)
	require.NoError(t, err)

	properties := valueobjects.NewRelationProperties(valueobjects.RelationPropertiesParams{
		SourceID:     sourceID,
		TargetID:     targetID,
		RelationType: relType,
		Name:         valueobjects.MustNewDescription(f.name),
		Description:  valueobjects.MustNewDescription(f.description),
	})

	relation, err := aggregates.NewComponentRelation(properties)
	require.NoError(t, err)

	return relation
}

func roundTripDeserialize(t *testing.T, events []domain.DomainEvent) []domain.DomainEvent {
	t.Helper()

	storedEvents := simulateRelationEventStoreRoundTrip(t, events)
	deserialized, err := relationEventDeserializers.Deserialize(storedEvents)
	require.NoError(t, err)

	return deserialized
}

func TestComponentRelationDeserializers_RoundTrip(t *testing.T) {
	original := newRelationFromFixture(t, relationFixture{
		sourceUUID:   "550e8400-e29b-41d4-a716-446655440000",
		targetUUID:   "550e8400-e29b-41d4-a716-446655440001",
		relationType: "Triggers",
		name:         "API Call",
		description:  "Service A calls Service B",
	})

	events := original.GetUncommittedChanges()
	require.Len(t, events, 1, "Expected 1 event: Created")

	deserializedEvents := roundTripDeserialize(t, events)
	require.Len(t, deserializedEvents, 1, "All events should be deserialized")

	loaded, err := aggregates.LoadComponentRelationFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.Equal(t, original.ID(), loaded.ID())
	assert.Equal(t, original.SourceComponentID().Value(), loaded.SourceComponentID().Value())
	assert.Equal(t, original.TargetComponentID().Value(), loaded.TargetComponentID().Value())
	assert.Equal(t, original.RelationType().Value(), loaded.RelationType().Value())
}

func TestComponentRelationDeserializers_RoundTripWithUpdate(t *testing.T) {
	original := newRelationFromFixture(t, relationFixture{
		sourceUUID:   "550e8400-e29b-41d4-a716-446655440002",
		targetUUID:   "550e8400-e29b-41d4-a716-446655440003",
		relationType: "Serves",
		name:         "Dependency",
		description:  "Initial description",
	})

	newName := valueobjects.MustNewDescription("Updated Dependency")
	newDescription := valueobjects.MustNewDescription("Updated description")
	_ = original.Update(newName, newDescription)

	events := original.GetUncommittedChanges()
	require.Len(t, events, 2, "Expected 2 events: Created, Updated")

	deserializedEvents := roundTripDeserialize(t, events)
	require.Len(t, deserializedEvents, 2, "All events should be deserialized")

	loaded, err := aggregates.LoadComponentRelationFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.Equal(t, newName.Value(), loaded.Name().Value())
	assert.Equal(t, newDescription.Value(), loaded.Description().Value())
}

func TestComponentRelationDeserializers_AllEventsCanBeDeserialized(t *testing.T) {
	relation := newRelationFromFixture(t, relationFixture{
		sourceUUID:   "550e8400-e29b-41d4-a716-446655440004",
		targetUUID:   "550e8400-e29b-41d4-a716-446655440005",
		relationType: "Triggers",
		name:         "Test Relation",
		description:  "Test description",
	})

	_ = relation.Update(
		valueobjects.MustNewDescription("Updated Name"),
		valueobjects.MustNewDescription("Updated description"),
	)

	events := relation.GetUncommittedChanges()
	require.Len(t, events, 2, "Expected 2 events: Created, Updated")

	deserializedEvents := roundTripDeserialize(t, events)
	require.Len(t, deserializedEvents, len(events),
		"All events should be deserialized - missing deserializer for one or more event types")

	for i, originalEvent := range events {
		assert.Equal(t, originalEvent.EventType(), deserializedEvents[i].EventType(),
			"Event type mismatch at index %d", i)
	}
}

type relationStoredEventWrapper struct {
	eventType string
	eventData map[string]interface{}
}

func (e *relationStoredEventWrapper) EventType() string                 { return e.eventType }
func (e *relationStoredEventWrapper) EventData() map[string]interface{} { return e.eventData }
func (e *relationStoredEventWrapper) AggregateID() string               { return "" }
func (e *relationStoredEventWrapper) OccurredAt() time.Time             { return time.Time{} }

func simulateRelationEventStoreRoundTrip(t *testing.T, events []domain.DomainEvent) []domain.DomainEvent {
	t.Helper()

	result := make([]domain.DomainEvent, len(events))

	for i, event := range events {
		jsonBytes, err := json.Marshal(event.EventData())
		require.NoError(t, err, "Failed to serialize event: %s", event.EventType())

		var data map[string]interface{}
		err = json.Unmarshal(jsonBytes, &data)
		require.NoError(t, err, "Failed to unmarshal JSON for event: %s", event.EventType())

		result[i] = &relationStoredEventWrapper{
			eventType: event.EventType(),
			eventData: data,
		}
	}

	return result
}
