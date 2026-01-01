package repositories

import (
	"encoding/json"
	"testing"
	"time"

	"easi/backend/internal/metamodel/domain/aggregates"
	"easi/backend/internal/metamodel/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetaModelConfigurationDeserializers_RoundTrip(t *testing.T) {
	tenantID, _ := sharedvo.NewTenantID("tenant-123")
	userEmail, _ := valueobjects.NewUserEmail("admin@example.com")

	original, err := aggregates.NewMetaModelConfiguration(tenantID, userEmail)
	require.NoError(t, err)

	events := original.GetUncommittedChanges()
	require.Len(t, events, 1, "Expected 1 event: Created")

	storedEvents := simulateMetaModelEventStoreRoundTrip(t, events)
	deserializedEvents := metaModelEventDeserializers.Deserialize(storedEvents)

	require.Len(t, deserializedEvents, 1, "All events should be deserialized")

	loaded, err := aggregates.LoadMetaModelConfigurationFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.Equal(t, original.ID(), loaded.ID())
	assert.Equal(t, original.TenantID().Value(), loaded.TenantID().Value())
}

func TestMetaModelConfigurationDeserializers_RoundTripWithMaturityUpdate(t *testing.T) {
	tenantID, _ := sharedvo.NewTenantID("tenant-456")
	userEmail, _ := valueobjects.NewUserEmail("admin@example.com")

	original, err := aggregates.NewMetaModelConfiguration(tenantID, userEmail)
	require.NoError(t, err)

	newConfig := original.MaturityScaleConfig()
	modifiedBy, _ := valueobjects.NewUserEmail("modifier@example.com")
	_ = original.UpdateMaturityScale(newConfig, modifiedBy)

	events := original.GetUncommittedChanges()
	require.Len(t, events, 2, "Expected 2 events: Created, Updated")

	storedEvents := simulateMetaModelEventStoreRoundTrip(t, events)
	deserializedEvents := metaModelEventDeserializers.Deserialize(storedEvents)

	require.Len(t, deserializedEvents, 2, "All events should be deserialized")

	_, err = aggregates.LoadMetaModelConfigurationFromHistory(deserializedEvents)
	require.NoError(t, err)
}

func TestMetaModelConfigurationDeserializers_RoundTripWithPillarChanges(t *testing.T) {
	tenantID, _ := sharedvo.NewTenantID("tenant-789")
	userEmail, _ := valueobjects.NewUserEmail("admin@example.com")

	original, err := aggregates.NewMetaModelConfiguration(tenantID, userEmail)
	require.NoError(t, err)

	pillarName, _ := valueobjects.NewPillarName("Innovation")
	pillarDesc, _ := valueobjects.NewPillarDescription("Focus on innovation")
	_ = original.AddStrategyPillar(pillarName, pillarDesc, userEmail)

	pillars := original.StrategyPillarsConfig().Pillars()
	if len(pillars) > 0 {
		newName, _ := valueobjects.NewPillarName("Updated Pillar")
		newDesc, _ := valueobjects.NewPillarDescription("Updated description")
		_ = original.UpdateStrategyPillar(pillars[0].ID(), newName, newDesc, userEmail)
	}

	events := original.GetUncommittedChanges()
	require.GreaterOrEqual(t, len(events), 2, "Expected at least 2 events")

	storedEvents := simulateMetaModelEventStoreRoundTrip(t, events)
	deserializedEvents := metaModelEventDeserializers.Deserialize(storedEvents)

	require.Len(t, deserializedEvents, len(events), "All events should be deserialized")

	_, err = aggregates.LoadMetaModelConfigurationFromHistory(deserializedEvents)
	require.NoError(t, err)
}

func TestMetaModelConfigurationDeserializers_AllEventsCanBeDeserialized(t *testing.T) {
	tenantID, _ := sharedvo.NewTenantID("tenant-test")
	userEmail, _ := valueobjects.NewUserEmail("admin@example.com")

	config, err := aggregates.NewMetaModelConfiguration(tenantID, userEmail)
	require.NoError(t, err)

	newConfig := config.MaturityScaleConfig()
	_ = config.UpdateMaturityScale(newConfig, userEmail)
	_ = config.ResetToDefaults(userEmail)

	pillarName, _ := valueobjects.NewPillarName("Test Pillar")
	pillarDesc, _ := valueobjects.NewPillarDescription("Test description")
	_ = config.AddStrategyPillar(pillarName, pillarDesc, userEmail)

	pillars := config.StrategyPillarsConfig().Pillars()
	var addedPillarID valueobjects.StrategyPillarID
	for _, p := range pillars {
		if p.Name().Value() == "Test Pillar" {
			addedPillarID = p.ID()
			break
		}
	}

	if addedPillarID.Value() != "" {
		newName, _ := valueobjects.NewPillarName("Updated Name")
		newDesc, _ := valueobjects.NewPillarDescription("Updated desc")
		_ = config.UpdateStrategyPillar(addedPillarID, newName, newDesc, userEmail)
		_ = config.RemoveStrategyPillar(addedPillarID, userEmail)
	}

	events := config.GetUncommittedChanges()
	require.GreaterOrEqual(t, len(events), 3, "Expected at least 3 events")

	storedEvents := simulateMetaModelEventStoreRoundTrip(t, events)
	deserializedEvents := metaModelEventDeserializers.Deserialize(storedEvents)

	require.Len(t, deserializedEvents, len(events),
		"All events should be deserialized - missing deserializer for one or more event types")

	for i, originalEvent := range events {
		assert.Equal(t, originalEvent.EventType(), deserializedEvents[i].EventType(),
			"Event type mismatch at index %d", i)
	}
}

type metaModelStoredEventWrapper struct {
	eventType string
	eventData map[string]interface{}
}

func (e *metaModelStoredEventWrapper) EventType() string                 { return e.eventType }
func (e *metaModelStoredEventWrapper) EventData() map[string]interface{} { return e.eventData }
func (e *metaModelStoredEventWrapper) AggregateID() string               { return "" }
func (e *metaModelStoredEventWrapper) OccurredAt() time.Time             { return time.Time{} }

func simulateMetaModelEventStoreRoundTrip(t *testing.T, events []domain.DomainEvent) []domain.DomainEvent {
	t.Helper()

	result := make([]domain.DomainEvent, len(events))

	for i, event := range events {
		jsonBytes, err := json.Marshal(event.EventData())
		require.NoError(t, err, "Failed to serialize event: %s", event.EventType())

		var data map[string]interface{}
		err = json.Unmarshal(jsonBytes, &data)
		require.NoError(t, err, "Failed to unmarshal JSON for event: %s", event.EventType())

		result[i] = &metaModelStoredEventWrapper{
			eventType: event.EventType(),
			eventData: data,
		}
	}

	return result
}
