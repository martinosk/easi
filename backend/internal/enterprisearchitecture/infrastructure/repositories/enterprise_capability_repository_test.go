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

func TestEnterpriseCapabilityDeserializers_RoundTrip(t *testing.T) {
	original := createTestCapability(t)

	name2, _ := valueobjects.NewEnterpriseCapabilityName("Updated Capability")
	desc2 := valueobjects.MustNewDescription("Updated description")
	cat2, _ := valueobjects.NewCategory("Finance")
	_ = original.Update(name2, desc2, cat2)

	targetMaturity, _ := valueobjects.NewTargetMaturity(85)
	_ = original.SetTargetMaturity(targetMaturity)

	events := original.GetUncommittedChanges()
	require.Len(t, events, 3, "Expected 3 events: Created, Updated, TargetMaturitySet")

	storedEvents := simulateEventStoreRoundTrip(t, events)
	deserializedEvents, err := enterpriseCapabilityEventDeserializers.Deserialize(storedEvents)
	require.NoError(t, err)

	require.Len(t, deserializedEvents, len(events), "All events should be deserialized")

	loaded, err := aggregates.LoadEnterpriseCapabilityFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.Equal(t, original.ID(), loaded.ID())
	assert.Equal(t, original.Name().Value(), loaded.Name().Value())
	assert.Equal(t, original.Description().Value(), loaded.Description().Value())
	assert.Equal(t, original.Category().Value(), loaded.Category().Value())
	assert.Equal(t, original.IsActive(), loaded.IsActive())
	require.NotNil(t, loaded.TargetMaturity(), "TargetMaturity should not be nil after round-trip")
	assert.Equal(t, original.TargetMaturity().Value(), loaded.TargetMaturity().Value())
}

func TestEnterpriseCapabilityDeserializers_RoundTripWithDelete(t *testing.T) {
	original := createTestCapability(t)
	_ = original.Delete()

	events := original.GetUncommittedChanges()
	storedEvents := simulateEventStoreRoundTrip(t, events)
	deserializedEvents, err := enterpriseCapabilityEventDeserializers.Deserialize(storedEvents)
	require.NoError(t, err)

	loaded, err := aggregates.LoadEnterpriseCapabilityFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.False(t, loaded.IsActive())
}

func TestEnterpriseCapabilityDeserializers_AllEventsCanBeDeserialized(t *testing.T) {
	capability := createTestCapability(t)

	name2, _ := valueobjects.NewEnterpriseCapabilityName("Updated Name")
	desc2 := valueobjects.MustNewDescription("Updated description")
	cat2, _ := valueobjects.NewCategory("Updated Category")
	_ = capability.Update(name2, desc2, cat2)

	targetMaturity, _ := valueobjects.NewTargetMaturity(75)
	_ = capability.SetTargetMaturity(targetMaturity)

	_ = capability.Delete()

	events := capability.GetUncommittedChanges()
	require.Len(t, events, 4, "Expected 4 events: Created, Updated, TargetMaturitySet, Deleted")

	storedEvents := simulateEventStoreRoundTrip(t, events)
	deserializedEvents, err := enterpriseCapabilityEventDeserializers.Deserialize(storedEvents)
	require.NoError(t, err)

	require.Len(t, deserializedEvents, 4,
		"All 4 events should be deserialized - missing deserializer for one or more event types")

	for i, original := range events {
		assert.Equal(t, original.EventType(), deserializedEvents[i].EventType(),
			"Event type mismatch at index %d", i)
	}
}

func createTestCapability(t *testing.T) *aggregates.EnterpriseCapability {
	t.Helper()

	name, err := valueobjects.NewEnterpriseCapabilityName("Test Capability")
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Test description")
	category, _ := valueobjects.NewCategory("Test Category")

	capability, err := aggregates.NewEnterpriseCapability(name, description, category)
	require.NoError(t, err)

	return capability
}

func simulateEventStoreRoundTrip(t *testing.T, events []domain.DomainEvent) []domain.DomainEvent {
	t.Helper()

	result := make([]domain.DomainEvent, len(events))

	for i, event := range events {
		jsonBytes, err := json.Marshal(event.EventData())
		require.NoError(t, err, "Failed to serialize event: %s", event.EventType())

		var data map[string]interface{}
		err = json.Unmarshal(jsonBytes, &data)
		require.NoError(t, err, "Failed to unmarshal JSON for event: %s", event.EventType())

		result[i] = &storedEventWrapper{
			eventType: event.EventType(),
			eventData: data,
		}
	}

	return result
}

type storedEventWrapper struct {
	eventType string
	eventData map[string]interface{}
}

func (e *storedEventWrapper) EventType() string                 { return e.eventType }
func (e *storedEventWrapper) EventData() map[string]interface{} { return e.eventData }
func (e *storedEventWrapper) AggregateID() string               { return "" }
func (e *storedEventWrapper) OccurredAt() time.Time             { return time.Time{} }
