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

func TestComponentRelationDeserializers_RoundTrip(t *testing.T) {
	sourceID, _ := valueobjects.NewComponentIDFromString("550e8400-e29b-41d4-a716-446655440000")
	targetID, _ := valueobjects.NewComponentIDFromString("550e8400-e29b-41d4-a716-446655440001")
	relationType, _ := valueobjects.NewRelationType("uses")
	name := valueobjects.MustNewDescription("API Call")
	description := valueobjects.MustNewDescription("Service A calls Service B")

	properties := valueobjects.NewRelationProperties(valueobjects.RelationPropertiesParams{
		SourceID:     sourceID,
		TargetID:     targetID,
		RelationType: relationType,
		Name:         name,
		Description:  description,
	})

	original, err := aggregates.NewComponentRelation(properties)
	require.NoError(t, err)

	events := original.GetUncommittedChanges()
	require.Len(t, events, 1, "Expected 1 event: Created")

	storedEvents := simulateRelationEventStoreRoundTrip(t, events)
	deserializedEvents := relationEventDeserializers.Deserialize(storedEvents)

	require.Len(t, deserializedEvents, 1, "All events should be deserialized")

	loaded, err := aggregates.LoadComponentRelationFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.Equal(t, original.ID(), loaded.ID())
	assert.Equal(t, original.SourceComponentID().Value(), loaded.SourceComponentID().Value())
	assert.Equal(t, original.TargetComponentID().Value(), loaded.TargetComponentID().Value())
	assert.Equal(t, original.RelationType().Value(), loaded.RelationType().Value())
}

func TestComponentRelationDeserializers_RoundTripWithUpdate(t *testing.T) {
	sourceID, _ := valueobjects.NewComponentIDFromString("550e8400-e29b-41d4-a716-446655440002")
	targetID, _ := valueobjects.NewComponentIDFromString("550e8400-e29b-41d4-a716-446655440003")
	relationType, _ := valueobjects.NewRelationType("depends-on")
	name := valueobjects.MustNewDescription("Dependency")
	description := valueobjects.MustNewDescription("Initial description")

	properties := valueobjects.NewRelationProperties(valueobjects.RelationPropertiesParams{
		SourceID:     sourceID,
		TargetID:     targetID,
		RelationType: relationType,
		Name:         name,
		Description:  description,
	})

	original, err := aggregates.NewComponentRelation(properties)
	require.NoError(t, err)

	newName := valueobjects.MustNewDescription("Updated Dependency")
	newDescription := valueobjects.MustNewDescription("Updated description")
	_ = original.Update(newName, newDescription)

	events := original.GetUncommittedChanges()
	require.Len(t, events, 2, "Expected 2 events: Created, Updated")

	storedEvents := simulateRelationEventStoreRoundTrip(t, events)
	deserializedEvents := relationEventDeserializers.Deserialize(storedEvents)

	require.Len(t, deserializedEvents, 2, "All events should be deserialized")

	loaded, err := aggregates.LoadComponentRelationFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.Equal(t, newName.Value(), loaded.Name().Value())
	assert.Equal(t, newDescription.Value(), loaded.Description().Value())
}

func TestComponentRelationDeserializers_AllEventsCanBeDeserialized(t *testing.T) {
	sourceID, _ := valueobjects.NewComponentIDFromString("550e8400-e29b-41d4-a716-446655440004")
	targetID, _ := valueobjects.NewComponentIDFromString("550e8400-e29b-41d4-a716-446655440005")
	relationType, _ := valueobjects.NewRelationType("uses")
	name := valueobjects.MustNewDescription("Test Relation")
	description := valueobjects.MustNewDescription("Test description")

	properties := valueobjects.NewRelationProperties(valueobjects.RelationPropertiesParams{
		SourceID:     sourceID,
		TargetID:     targetID,
		RelationType: relationType,
		Name:         name,
		Description:  description,
	})

	relation, err := aggregates.NewComponentRelation(properties)
	require.NoError(t, err)

	newName := valueobjects.MustNewDescription("Updated Name")
	newDescription := valueobjects.MustNewDescription("Updated description")
	_ = relation.Update(newName, newDescription)

	events := relation.GetUncommittedChanges()
	require.Len(t, events, 2, "Expected 2 events: Created, Updated")

	storedEvents := simulateRelationEventStoreRoundTrip(t, events)
	deserializedEvents := relationEventDeserializers.Deserialize(storedEvents)

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
