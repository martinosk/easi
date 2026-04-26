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

const testAdminEmail = "admin@example.com"

func TestMetaModelConfigurationDeserializers_RoundTrip(t *testing.T) {
	original, _ := newTestConfig(t, "tenant-123")

	loaded := roundTripAndLoad(t, original, 1)

	assert.Equal(t, original.ID(), loaded.ID())
	assert.Equal(t, original.TenantID().Value(), loaded.TenantID().Value())
}

func TestMetaModelConfigurationDeserializers_RoundTripWithMaturityUpdate(t *testing.T) {
	original, _ := newTestConfig(t, "tenant-456")

	modifiedBy, err := valueobjects.NewUserEmail("modifier@example.com")
	require.NoError(t, err)
	require.NoError(t, original.UpdateMaturityScale(original.MaturityScaleConfig(), modifiedBy))

	roundTripAndLoad(t, original, 2)
}

func TestMetaModelConfigurationDeserializers_RoundTripWithPillarChanges(t *testing.T) {
	original, userEmail := newTestConfig(t, "tenant-789")

	addPillar(t, original, "Innovation", "Focus on innovation")
	pillar := pickPillar(t, original.StrategyPillarsConfig().Pillars(), byPillarName("Innovation"))

	updatedName, err := valueobjects.NewPillarName("Updated Pillar")
	require.NoError(t, err)
	updatedDesc, err := valueobjects.NewPillarDescription("Updated description")
	require.NoError(t, err)
	require.NoError(t, original.UpdateStrategyPillar(pillar.ID(), updatedName, updatedDesc, userEmail))

	events := original.GetUncommittedChanges()
	require.GreaterOrEqual(t, len(events), 2)

	roundTripAndLoad(t, original, len(events))
}

func TestMetaModelConfigurationDeserializers_AllEventsCanBeDeserialized(t *testing.T) {
	config, userEmail := newTestConfig(t, "tenant-test")

	require.NoError(t, config.UpdateMaturityScale(config.MaturityScaleConfig(), userEmail))
	require.NoError(t, config.ResetToDefaults(userEmail))

	addPillar(t, config, "Test Pillar", "Test description")
	pillar := pickPillar(t, config.StrategyPillarsConfig().Pillars(), byPillarName("Test Pillar"))

	updatedName, err := valueobjects.NewPillarName("Updated Name")
	require.NoError(t, err)
	updatedDesc, err := valueobjects.NewPillarDescription("Updated desc")
	require.NoError(t, err)
	require.NoError(t, config.UpdateStrategyPillar(pillar.ID(), updatedName, updatedDesc, userEmail))

	require.NoError(t, config.UpdatePillarFitConfiguration(pillar.ID(), newTestFitConfig(t, "Test criteria"), userEmail))
	require.NoError(t, config.RemoveStrategyPillar(pillar.ID(), userEmail))

	events := config.GetUncommittedChanges()
	requireEventTypesPresent(t, events,
		"MetaModelConfigurationCreated",
		"MaturityScaleConfigUpdated",
		"MaturityScaleConfigReset",
		"StrategyPillarAdded",
		"StrategyPillarUpdated",
		"PillarFitConfigurationUpdated",
		"StrategyPillarRemoved",
	)

	loaded := roundTripAndLoad(t, config, len(events))
	assert.Equal(t, config.Version(), loaded.Version(),
		"Loaded aggregate version must match original after deserializing all events")
}

func TestMetaModelConfigurationDeserializers_PillarFitConfigurationUpdated(t *testing.T) {
	config, userEmail := newTestConfig(t, "tenant-fit-config")

	addPillar(t, config, "Cloud Native", "Cloud native capabilities")
	pillar := pickPillar(t, config.StrategyPillarsConfig().Pillars(), byPillarName("Cloud Native"))

	require.NoError(t, config.UpdatePillarFitConfiguration(pillar.ID(),
		newTestFitConfig(t, "Containerization, Kubernetes, CI/CD"), userEmail))

	events := config.GetUncommittedChanges()
	require.Len(t, events, 3, "Expected 3 events: Created, PillarAdded, FitConfigUpdated")
	requireEventTypesPresent(t, events, "PillarFitConfigurationUpdated")

	loaded := roundTripAndLoad(t, config, len(events))
	assert.Equal(t, config.Version(), loaded.Version(),
		"Loaded aggregate version must match original - version mismatch causes optimistic locking failures")

	loadedPillar := pickPillar(t, loaded.StrategyPillarsConfig().Pillars(), byPillarID(pillar.ID()))
	assert.True(t, loadedPillar.FitScoringEnabled(), "Fit scoring should be enabled after rehydration")
	assert.Equal(t, "Containerization, Kubernetes, CI/CD", loadedPillar.FitCriteria().Value(),
		"Fit criteria should be preserved after rehydration")
}

func newTestConfig(t *testing.T, tenantIDStr string) (*aggregates.MetaModelConfiguration, valueobjects.UserEmail) {
	t.Helper()
	tenantID, err := sharedvo.NewTenantID(tenantIDStr)
	require.NoError(t, err)
	userEmail, err := valueobjects.NewUserEmail(testAdminEmail)
	require.NoError(t, err)
	config, err := aggregates.NewMetaModelConfiguration(tenantID, userEmail)
	require.NoError(t, err)
	return config, userEmail
}

func addPillar(t *testing.T, config *aggregates.MetaModelConfiguration, name, description string) {
	t.Helper()
	pillarName, err := valueobjects.NewPillarName(name)
	require.NoError(t, err)
	pillarDesc, err := valueobjects.NewPillarDescription(description)
	require.NoError(t, err)
	userEmail, err := valueobjects.NewUserEmail(testAdminEmail)
	require.NoError(t, err)
	require.NoError(t, config.AddStrategyPillar(pillarName, pillarDesc, userEmail))
}

func newTestFitConfig(t *testing.T, criteria string) valueobjects.FitConfigurationParams {
	t.Helper()
	fitCriteria, err := valueobjects.NewFitCriteria(criteria)
	require.NoError(t, err)
	fitType, err := valueobjects.NewFitType("TECHNICAL")
	require.NoError(t, err)
	return valueobjects.NewFitConfigurationParams(true, fitCriteria, fitType)
}

func pickPillar(t *testing.T, pillars []valueobjects.StrategyPillar, match func(valueobjects.StrategyPillar) bool) valueobjects.StrategyPillar {
	t.Helper()
	for _, p := range pillars {
		if match(p) {
			return p
		}
	}
	require.Fail(t, "pillar not found")
	return valueobjects.StrategyPillar{}
}

func byPillarName(name string) func(valueobjects.StrategyPillar) bool {
	return func(p valueobjects.StrategyPillar) bool { return p.Name().Value() == name }
}

func byPillarID(id valueobjects.StrategyPillarID) func(valueobjects.StrategyPillar) bool {
	return func(p valueobjects.StrategyPillar) bool { return p.ID().Value() == id.Value() }
}

func requireEventTypesPresent(t *testing.T, events []domain.DomainEvent, expected ...string) {
	t.Helper()
	seen := make(map[string]bool, len(events))
	for _, e := range events {
		seen[e.EventType()] = true
	}
	for _, eventType := range expected {
		require.Truef(t, seen[eventType], "Expected event type %q in events", eventType)
	}
}

func roundTripAndLoad(t *testing.T, config *aggregates.MetaModelConfiguration, expectedEventCount int) *aggregates.MetaModelConfiguration {
	t.Helper()
	events := config.GetUncommittedChanges()
	require.Len(t, events, expectedEventCount)

	storedEvents := simulateMetaModelEventStoreRoundTrip(t, events)
	deserializedEvents, err := metaModelEventDeserializers.Deserialize(storedEvents)
	require.NoError(t, err)
	require.Len(t, deserializedEvents, expectedEventCount,
		"All events should be deserialized - missing deserializer for one or more event types")

	for i, originalEvent := range events {
		assert.Equal(t, originalEvent.EventType(), deserializedEvents[i].EventType(),
			"Event type mismatch at index %d", i)
	}

	loaded, err := aggregates.LoadMetaModelConfigurationFromHistory(deserializedEvents)
	require.NoError(t, err)
	return loaded
}

type metaModelStoredEventWrapper struct {
	eventType string
	eventData map[string]any
}

func (e *metaModelStoredEventWrapper) EventType() string         { return e.eventType }
func (e *metaModelStoredEventWrapper) EventData() map[string]any { return e.eventData }
func (e *metaModelStoredEventWrapper) AggregateID() string       { return "" }
func (e *metaModelStoredEventWrapper) OccurredAt() time.Time     { return time.Time{} }

func simulateMetaModelEventStoreRoundTrip(t *testing.T, events []domain.DomainEvent) []domain.DomainEvent {
	t.Helper()

	result := make([]domain.DomainEvent, len(events))

	for i, event := range events {
		jsonBytes, err := json.Marshal(event.EventData())
		require.NoError(t, err, "Failed to serialize event: %s", event.EventType())

		var data map[string]any
		err = json.Unmarshal(jsonBytes, &data)
		require.NoError(t, err, "Failed to unmarshal JSON for event: %s", event.EventType())

		result[i] = &metaModelStoredEventWrapper{
			eventType: event.EventType(),
			eventData: data,
		}
	}

	return result
}
