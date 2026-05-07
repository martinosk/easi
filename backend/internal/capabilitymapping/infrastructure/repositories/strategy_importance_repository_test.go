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

type strategyImportanceFixture struct {
	pillarName     string
	importanceRank int
	rationale      string
}

func newStrategyImportanceFromFixture(t *testing.T, f strategyImportanceFixture) *aggregates.StrategyImportance {
	t.Helper()

	importance, err := valueobjects.NewImportance(f.importanceRank)
	require.NoError(t, err)
	rationale, err := valueobjects.NewRationale(f.rationale)
	require.NoError(t, err)

	original, err := aggregates.SetStrategyImportance(aggregates.NewImportanceParams{
		BusinessDomainID: valueobjects.NewBusinessDomainID(),
		CapabilityID:     valueobjects.NewCapabilityID(),
		PillarID:         valueobjects.NewPillarID(),
		PillarName:       f.pillarName,
		Importance:       importance,
		Rationale:        rationale,
	})
	require.NoError(t, err)

	return original
}

func roundTripDeserializeStrategyImportance(t *testing.T, events []domain.DomainEvent) []domain.DomainEvent {
	t.Helper()

	storedEvents := simulateStrategyImportanceEventStoreRoundTrip(t, events)
	deserialized, err := strategyImportanceEventDeserializers.Deserialize(storedEvents)
	require.NoError(t, err)

	return deserialized
}

func TestStrategyImportanceDeserializers_RoundTrip(t *testing.T) {
	original := newStrategyImportanceFromFixture(t, strategyImportanceFixture{
		pillarName:     "Growth",
		importanceRank: 4,
		rationale:      "Critical for growth",
	})

	events := original.GetUncommittedChanges()
	require.Len(t, events, 1, "Expected 1 event: Set")

	deserializedEvents := roundTripDeserializeStrategyImportance(t, events)
	require.Len(t, deserializedEvents, 1, "All events should be deserialized")

	loaded, err := aggregates.LoadStrategyImportanceFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.Equal(t, original.ID(), loaded.ID())
	assert.Equal(t, original.BusinessDomainID().Value(), loaded.BusinessDomainID().Value())
	assert.Equal(t, original.CapabilityID().Value(), loaded.CapabilityID().Value())
	assert.Equal(t, original.PillarID().Value(), loaded.PillarID().Value())
	assert.Equal(t, original.Importance().Value(), loaded.Importance().Value())
	assert.Equal(t, original.Rationale().Value(), loaded.Rationale().Value())
}

func TestStrategyImportanceDeserializers_RoundTripWithUpdate(t *testing.T) {
	original := newStrategyImportanceFromFixture(t, strategyImportanceFixture{
		pillarName:     "Innovation",
		importanceRank: 3,
		rationale:      "Initial rationale",
	})

	newImportance, _ := valueobjects.NewImportance(5)
	newRationale, _ := valueobjects.NewRationale("Updated rationale")
	_ = original.Update(newImportance, newRationale)

	events := original.GetUncommittedChanges()
	require.Len(t, events, 2, "Expected 2 events: Set, Updated")

	deserializedEvents := roundTripDeserializeStrategyImportance(t, events)
	require.Len(t, deserializedEvents, 2, "All events should be deserialized")

	loaded, err := aggregates.LoadStrategyImportanceFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.Equal(t, newImportance.Value(), loaded.Importance().Value())
	assert.Equal(t, newRationale.Value(), loaded.Rationale().Value())
}

func TestStrategyImportanceDeserializers_AllEventsCanBeDeserialized(t *testing.T) {
	strategyImportance := newStrategyImportanceFromFixture(t, strategyImportanceFixture{
		pillarName:     "Efficiency",
		importanceRank: 2,
		rationale:      "Test rationale",
	})

	newImportance, _ := valueobjects.NewImportance(4)
	newRationale, _ := valueobjects.NewRationale("Updated")
	_ = strategyImportance.Update(newImportance, newRationale)

	_ = strategyImportance.Remove()

	events := strategyImportance.GetUncommittedChanges()
	require.Len(t, events, 3, "Expected 3 events: Set, Updated, Removed")

	deserializedEvents := roundTripDeserializeStrategyImportance(t, events)
	require.Len(t, deserializedEvents, len(events),
		"All events should be deserialized - missing deserializer for one or more event types")

	for i, originalEvent := range events {
		assert.Equal(t, originalEvent.EventType(), deserializedEvents[i].EventType(),
			"Event type mismatch at index %d", i)
	}
}

type strategyImportanceStoredEventWrapper struct {
	eventType string
	eventData map[string]interface{}
}

func (e *strategyImportanceStoredEventWrapper) EventType() string                 { return e.eventType }
func (e *strategyImportanceStoredEventWrapper) EventData() map[string]interface{} { return e.eventData }
func (e *strategyImportanceStoredEventWrapper) AggregateID() string               { return "" }
func (e *strategyImportanceStoredEventWrapper) OccurredAt() time.Time             { return time.Time{} }

func simulateStrategyImportanceEventStoreRoundTrip(t *testing.T, events []domain.DomainEvent) []domain.DomainEvent {
	t.Helper()

	result := make([]domain.DomainEvent, len(events))

	for i, event := range events {
		jsonBytes, err := json.Marshal(event.EventData())
		require.NoError(t, err, "Failed to serialize event: %s", event.EventType())

		var data map[string]interface{}
		err = json.Unmarshal(jsonBytes, &data)
		require.NoError(t, err, "Failed to unmarshal JSON for event: %s", event.EventType())

		result[i] = &strategyImportanceStoredEventWrapper{
			eventType: event.EventType(),
			eventData: data,
		}
	}

	return result
}
