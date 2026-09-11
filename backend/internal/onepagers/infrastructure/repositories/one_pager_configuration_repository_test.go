package repositories

import (
	"encoding/json"
	"testing"
	"time"

	"easi/backend/internal/onepagers/domain/aggregates"
	opevents "easi/backend/internal/onepagers/domain/events"
	"easi/backend/internal/onepagers/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAdminEmail = "admin@example.com"

func newTestConfig(t *testing.T) (*aggregates.OnePagerConfiguration, valueobjects.UserEmail) {
	t.Helper()
	tenantID, err := sharedvo.NewTenantID("tenant-123")
	require.NoError(t, err)
	subjectType, err := valueobjects.NewSubjectType("application")
	require.NoError(t, err)
	userEmail, err := valueobjects.NewUserEmail(testAdminEmail)
	require.NoError(t, err)
	config, err := aggregates.NewOnePagerConfiguration(tenantID, subjectType, userEmail)
	require.NoError(t, err)
	return config, userEmail
}

func TestOnePagerConfigurationDeserializers_PresentationEventsRoundTrip(t *testing.T) {
	config, userEmail := newTestConfig(t)
	fieldID := valueobjects.NewFieldID()

	require.NoError(t, config.IncludeCustomField(fieldID, userEmail))
	require.NoError(t, config.ChangeCustomFieldRequirement(fieldID, true, userEmail))
	require.NoError(t, config.ExcludeCustomField(fieldID, userEmail))
	require.NoError(t, config.IncludeCustomField(fieldID, userEmail))
	require.NoError(t, config.ExcludeBuiltInField("experts", userEmail))
	require.NoError(t, config.IncludeBuiltInField("experts", userEmail))
	require.NoError(t, config.ChangeBuiltInFieldRequirement("experts", true, userEmail))
	require.NoError(t, config.ReorderFields(reverseOrder(config.DisplayOrder()), userEmail))

	events := config.GetUncommittedChanges()
	requireEventTypesPresent(t, events,
		"OnePagerConfigurationCreated",
		"CustomFieldIncluded",
		"CustomFieldRequirementChanged",
		"CustomFieldExcluded",
		"BuiltInFieldIncluded",
		"BuiltInFieldExcluded",
		"BuiltInFieldRequirementChanged",
		"OnePagerFieldsReordered",
	)

	loaded := roundTripAndLoad(t, events, len(events))

	assert.Equal(t, config.ID(), loaded.ID())
	assert.Equal(t, config.Version(), loaded.Version())
	assert.Equal(t, config.SubjectType().Value(), loaded.SubjectType().Value())
	assert.Equal(t, config.DisplayOrder(), loaded.DisplayOrder())
	assert.True(t, loaded.IsCustomFieldRequired(fieldID))
	assert.True(t, loaded.IsBuiltInRequired("experts"))
}

func legacyParams(configID string, version int) opevents.ModifyConfigurationParams {
	return opevents.ModifyConfigurationParams{ConfigID: configID, TenantID: "tenant-123", Version: version, ModifiedBy: testAdminEmail}
}

func TestOnePagerConfigurationDeserializers_LegacySchemaEventsStillReplay(t *testing.T) {
	config, _ := newTestConfig(t)
	fieldID := valueobjects.NewFieldID()
	optionID := valueobjects.NewOptionID().Value()
	min := 0.0
	history := append(config.GetUncommittedChanges(),
		opevents.NewCustomFieldDefined(legacyParams(config.ID(), 2), opevents.CustomFieldData{
			FieldID: fieldID.Value(), Name: "Hosting model", FieldType: "selection", Required: true,
			Options: []opevents.SelectionOptionData{{ID: optionID, Label: "Cloud", Active: true}},
		}),
		opevents.NewCustomFieldRenamed(legacyParams(config.ID(), 3), opevents.FieldRenameData{FieldID: fieldID.Value(), NewName: "Deployment"}),
		opevents.NewSelectionOptionAdded(legacyParams(config.ID(), 4), fieldID.Value(), valueobjects.NewOptionID().Value(), "Hybrid"),
		opevents.NewSelectionOptionRetired(legacyParams(config.ID(), 5), fieldID.Value(), optionID),
		opevents.NewNumberFieldBoundsChanged(legacyParams(config.ID(), 6), fieldID.Value(), &min, nil),
		opevents.NewCustomFieldRetired(legacyParams(config.ID(), 7), fieldID.Value()),
		opevents.NewCustomFieldReactivated(legacyParams(config.ID(), 8), fieldID.Value()),
	)

	loaded := roundTripAndLoad(t, history, len(history))

	assert.Equal(t, 8, loaded.Version())
	assert.True(t, loaded.IsCustomFieldIncluded(fieldID))
	assert.True(t, loaded.IsCustomFieldRequired(fieldID))
}

func TestOnePagerEventDeserializers_CoverEveryConfigurationEventType(t *testing.T) {
	for _, eventType := range opevents.ConfigurationEventTypes() {
		assert.Truef(t, onePagerEventDeserializers.HasDeserializerFor(eventType),
			"No deserializer registered for %q: the event store silently skips unknown event types, "+
				"so the aggregate reloads a version lower than the stored one and every subsequent "+
				"write fails with a concurrency conflict", eventType)
	}
}

func TestOnePagerFactsEventDeserializers_CoverEveryFactsEventType(t *testing.T) {
	for _, eventType := range opevents.FactsEventTypes() {
		assert.Truef(t, onePagerFactsEventDeserializers.HasDeserializerFor(eventType),
			"No deserializer registered for %q", eventType)
	}
}

func reverseOrder(order []valueobjects.FieldRef) []valueobjects.FieldRef {
	reversed := make([]valueobjects.FieldRef, len(order))
	for i, ref := range order {
		reversed[len(order)-1-i] = ref
	}
	return reversed
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

func roundTripAndLoad(t *testing.T, events []domain.DomainEvent, expectedEventCount int) *aggregates.OnePagerConfiguration {
	t.Helper()
	require.Len(t, events, expectedEventCount)

	storedEvents := simulateEventStoreRoundTrip(t, events)
	deserializedEvents, err := onePagerEventDeserializers.Deserialize(storedEvents)
	require.NoError(t, err)
	require.Len(t, deserializedEvents, expectedEventCount,
		"All events should be deserialized - missing deserializer for one or more event types")

	for i, originalEvent := range events {
		assert.Equal(t, originalEvent.EventType(), deserializedEvents[i].EventType(),
			"Event type mismatch at index %d", i)
	}

	loaded, err := aggregates.LoadOnePagerConfigurationFromHistory(deserializedEvents)
	require.NoError(t, err)
	return loaded
}

type storedEventWrapper struct {
	eventType string
	eventData map[string]any
}

func (e *storedEventWrapper) EventType() string         { return e.eventType }
func (e *storedEventWrapper) EventData() map[string]any { return e.eventData }
func (e *storedEventWrapper) AggregateID() string       { return "" }
func (e *storedEventWrapper) OccurredAt() time.Time     { return time.Time{} }

func simulateEventStoreRoundTrip(t *testing.T, events []domain.DomainEvent) []domain.DomainEvent {
	t.Helper()

	result := make([]domain.DomainEvent, len(events))
	for i, event := range events {
		jsonBytes, err := json.Marshal(event.EventData())
		require.NoError(t, err, "Failed to serialize event: %s", event.EventType())

		var data map[string]any
		err = json.Unmarshal(jsonBytes, &data)
		require.NoError(t, err, "Failed to unmarshal JSON for event: %s", event.EventType())

		result[i] = &storedEventWrapper{eventType: event.EventType(), eventData: data}
	}
	return result
}
