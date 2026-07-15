package repository

import (
	"testing"
	"time"

	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type knownEvent struct {
	Name string `json:"name"`
}

func (e knownEvent) EventType() string   { return "KnownEvent" }
func (e knownEvent) AggregateID() string { return "aggregate-1" }
func (e knownEvent) OccurredAt() time.Time {
	return time.Time{}
}

func (e knownEvent) EventData() map[string]interface{} {
	return map[string]interface{}{"name": e.Name}
}

type storedEvent struct {
	eventType   string
	aggregateID string
	data        map[string]interface{}
}

func (e storedEvent) EventType() string                 { return e.eventType }
func (e storedEvent) AggregateID() string               { return e.aggregateID }
func (e storedEvent) OccurredAt() time.Time             { return time.Time{} }
func (e storedEvent) EventData() map[string]interface{} { return e.data }

func testDeserializers() EventDeserializers {
	return NewEventDeserializers(map[string]EventDeserializerFunc{
		"KnownEvent": JSONDeserializer[knownEvent],
	})
}

func TestDeserialize_UnknownEventTypeReturnsError(t *testing.T) {
	stored := []domain.DomainEvent{
		storedEvent{eventType: "KnownEvent", aggregateID: "aggregate-1", data: map[string]interface{}{"name": "first"}},
		storedEvent{eventType: "MysteryEvent", aggregateID: "aggregate-1", data: map[string]interface{}{}},
	}

	result, err := testDeserializers().Deserialize(stored)

	require.Error(t, err, "An unregistered event type must fail loudly, never be skipped")
	assert.Nil(t, result)

	var unknownErr *UnknownEventTypeError
	require.ErrorAs(t, err, &unknownErr)
	assert.Equal(t, "MysteryEvent", unknownErr.EventType)
	assert.Equal(t, "aggregate-1", unknownErr.AggregateID)
	assert.Equal(t, 2, unknownErr.SequenceNumber)
}

func TestDeserialize_UnknownEventTypeNeverSilentlyDropsEvents(t *testing.T) {
	stored := []domain.DomainEvent{
		storedEvent{eventType: "KnownEvent", aggregateID: "aggregate-1", data: map[string]interface{}{"name": "first"}},
		storedEvent{eventType: "MysteryEvent", aggregateID: "aggregate-1", data: map[string]interface{}{}},
		storedEvent{eventType: "KnownEvent", aggregateID: "aggregate-1", data: map[string]interface{}{"name": "third"}},
	}

	_, err := testDeserializers().Deserialize(stored)

	require.Error(t, err,
		"Skipping the unknown event would load 2 of 3 events, undercounting the aggregate version "+
			"and failing every later write with a spurious concurrency conflict")
}

func TestDeserialize_AllKnownEventTypesSucceed(t *testing.T) {
	stored := []domain.DomainEvent{
		storedEvent{eventType: "KnownEvent", aggregateID: "aggregate-1", data: map[string]interface{}{"name": "first"}},
		storedEvent{eventType: "KnownEvent", aggregateID: "aggregate-1", data: map[string]interface{}{"name": "second"}},
	}

	result, err := testDeserializers().Deserialize(stored)

	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "first", result[0].(knownEvent).Name)
	assert.Equal(t, "second", result[1].(knownEvent).Name)
}

func TestHasDeserializerFor(t *testing.T) {
	deserializers := testDeserializers()

	assert.True(t, deserializers.HasDeserializerFor("KnownEvent"))
	assert.False(t, deserializers.HasDeserializerFor("MysteryEvent"))
}
