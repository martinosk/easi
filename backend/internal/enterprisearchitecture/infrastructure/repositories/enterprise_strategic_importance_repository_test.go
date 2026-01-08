package repositories

import (
	"encoding/json"
	"testing"
	"time"

	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnterpriseStrategicImportanceDeserializers_RoundTrip(t *testing.T) {
	enterpriseCapabilityID := valueobjects.NewEnterpriseCapabilityID()
	pillarID := valueobjects.NewPillarID()
	importance, _ := valueobjects.NewImportance(4)
	rationale, _ := valueobjects.NewRationale("Critical for growth")

	original, err := aggregates.SetEnterpriseStrategicImportance(aggregates.NewEnterpriseImportanceParams{
		EnterpriseCapabilityID: enterpriseCapabilityID,
		PillarID:               pillarID,
		PillarName:             "Growth",
		Importance:             importance,
		Rationale:              rationale,
	})
	require.NoError(t, err)

	events := original.GetUncommittedChanges()
	require.Len(t, events, 1, "Expected 1 event: Set")

	storedEvents := simulateImportanceEventStoreRoundTrip(t, events)
	deserializedEvents, err := enterpriseStrategicImportanceEventDeserializers.Deserialize(storedEvents)
	require.NoError(t, err)

	require.Len(t, deserializedEvents, 1, "All events should be deserialized")

	loaded, err := aggregates.LoadEnterpriseStrategicImportanceFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.Equal(t, original.ID(), loaded.ID())
	assert.Equal(t, original.EnterpriseCapabilityID().Value(), loaded.EnterpriseCapabilityID().Value())
	assert.Equal(t, original.PillarID().Value(), loaded.PillarID().Value())
	assert.Equal(t, original.Importance().Value(), loaded.Importance().Value())
	assert.Equal(t, original.Rationale().Value(), loaded.Rationale().Value())
}

func TestEnterpriseStrategicImportanceDeserializers_RoundTripWithUpdate(t *testing.T) {
	enterpriseCapabilityID := valueobjects.NewEnterpriseCapabilityID()
	pillarID := valueobjects.NewPillarID()
	importance, _ := valueobjects.NewImportance(3)
	rationale, _ := valueobjects.NewRationale("Initial rationale")

	original, err := aggregates.SetEnterpriseStrategicImportance(aggregates.NewEnterpriseImportanceParams{
		EnterpriseCapabilityID: enterpriseCapabilityID,
		PillarID:               pillarID,
		PillarName:             "Innovation",
		Importance:             importance,
		Rationale:              rationale,
	})
	require.NoError(t, err)

	newImportance, _ := valueobjects.NewImportance(5)
	newRationale, _ := valueobjects.NewRationale("Updated rationale")
	_ = original.Update(newImportance, newRationale)

	events := original.GetUncommittedChanges()
	require.Len(t, events, 2, "Expected 2 events: Set, Updated")

	storedEvents := simulateImportanceEventStoreRoundTrip(t, events)
	deserializedEvents, err := enterpriseStrategicImportanceEventDeserializers.Deserialize(storedEvents)
	require.NoError(t, err)

	require.Len(t, deserializedEvents, 2, "All events should be deserialized")

	loaded, err := aggregates.LoadEnterpriseStrategicImportanceFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.Equal(t, newImportance.Value(), loaded.Importance().Value())
	assert.Equal(t, newRationale.Value(), loaded.Rationale().Value())
}

func TestEnterpriseStrategicImportanceDeserializers_AllEventsCanBeDeserialized(t *testing.T) {
	enterpriseCapabilityID := valueobjects.NewEnterpriseCapabilityID()
	pillarID := valueobjects.NewPillarID()
	importance, _ := valueobjects.NewImportance(2)
	rationale, _ := valueobjects.NewRationale("Test rationale")

	original, err := aggregates.SetEnterpriseStrategicImportance(aggregates.NewEnterpriseImportanceParams{
		EnterpriseCapabilityID: enterpriseCapabilityID,
		PillarID:               pillarID,
		PillarName:             "Efficiency",
		Importance:             importance,
		Rationale:              rationale,
	})
	require.NoError(t, err)

	newImportance, _ := valueobjects.NewImportance(4)
	newRationale, _ := valueobjects.NewRationale("Updated")
	_ = original.Update(newImportance, newRationale)
	_ = original.Remove()

	events := original.GetUncommittedChanges()
	require.Len(t, events, 3, "Expected 3 events: Set, Updated, Removed")

	storedEvents := simulateImportanceEventStoreRoundTrip(t, events)
	deserializedEvents, err := enterpriseStrategicImportanceEventDeserializers.Deserialize(storedEvents)
	require.NoError(t, err)

	require.Len(t, deserializedEvents, len(events),
		"All events should be deserialized - missing deserializer for one or more event types")

	for i, originalEvent := range events {
		assert.Equal(t, originalEvent.EventType(), deserializedEvents[i].EventType(),
			"Event type mismatch at index %d", i)
	}
}

type importanceStoredEventWrapper struct {
	eventType string
	eventData map[string]interface{}
}

func (e *importanceStoredEventWrapper) EventType() string                 { return e.eventType }
func (e *importanceStoredEventWrapper) EventData() map[string]interface{} { return e.eventData }
func (e *importanceStoredEventWrapper) AggregateID() string               { return "" }
func (e *importanceStoredEventWrapper) OccurredAt() time.Time             { return time.Time{} }

func simulateImportanceEventStoreRoundTrip(t *testing.T, events []domain.DomainEvent) []domain.DomainEvent {
	t.Helper()

	result := make([]domain.DomainEvent, len(events))

	for i, event := range events {
		jsonBytes, err := json.Marshal(event.EventData())
		require.NoError(t, err, "Failed to serialize event: %s", event.EventType())

		var data map[string]interface{}
		err = json.Unmarshal(jsonBytes, &data)
		require.NoError(t, err, "Failed to unmarshal JSON for event: %s", event.EventType())

		result[i] = &importanceStoredEventWrapper{
			eventType: event.EventType(),
			eventData: data,
		}
	}

	return result
}
