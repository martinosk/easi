package repositories

import (
	"encoding/json"
	"testing"
	"time"

	"easi/backend/internal/onepagers/domain/aggregates"
	"easi/backend/internal/onepagers/domain/events"
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

func mustName(t *testing.T, v string) valueobjects.FieldName {
	t.Helper()
	name, err := valueobjects.NewFieldName(v)
	require.NoError(t, err)
	return name
}

func mustType(t *testing.T, v string) valueobjects.FieldType {
	t.Helper()
	ft, err := valueobjects.NewFieldType(v)
	require.NoError(t, err)
	return ft
}

func mustLabel(t *testing.T, v string) valueobjects.OptionLabel {
	t.Helper()
	label, err := valueobjects.NewOptionLabel(v)
	require.NoError(t, err)
	return label
}

func TestOnePagerConfigurationDeserializers_AllEventsRoundTrip(t *testing.T) {
	config, userEmail := newTestConfig(t)

	textID, err := config.DefineCustomField(aggregates.DefineCustomFieldParams{
		Name: mustName(t, "Business summary"),
		Type: mustType(t, "text"),
	}, userEmail)
	require.NoError(t, err)

	maturityID, err := config.DefineCustomField(aggregates.DefineCustomFieldParams{
		Name: mustName(t, "Maturity score"),
		Type: mustType(t, "number"),
	}, userEmail)
	require.NoError(t, err)
	minBound := 0.0
	maxBound := 5.0
	require.NoError(t, config.SetNumberFieldBounds(maturityID, &minBound, &maxBound, userEmail))

	selectionID, err := config.DefineCustomField(aggregates.DefineCustomFieldParams{
		Name:         mustName(t, "Hosting model"),
		Type:         mustType(t, "selection"),
		OptionLabels: []valueobjects.OptionLabel{mustLabel(t, "On-prem"), mustLabel(t, "Cloud")},
	}, userEmail)
	require.NoError(t, err)

	help, err := valueobjects.NewHelpText("A short summary")
	require.NoError(t, err)
	require.NoError(t, config.RenameCustomField(aggregates.RenameCustomFieldParams{
		FieldID:  textID,
		Name:     mustName(t, "Summary"),
		HelpText: help,
	}, userEmail))
	require.NoError(t, config.ChangeCustomFieldRequirement(textID, true, userEmail))
	require.NoError(t, config.ExcludeBuiltInField("experts", userEmail))
	require.NoError(t, config.IncludeBuiltInField("experts", userEmail))
	require.NoError(t, config.ChangeBuiltInFieldRequirement("experts", true, userEmail))

	_, err = config.AddSelectionOption(selectionID, mustLabel(t, "Hybrid"), userEmail)
	require.NoError(t, err)
	field, found := config.CustomFieldByID(selectionID)
	require.True(t, found)
	require.NoError(t, config.RetireSelectionOption(selectionID, field.Options()[0].ID(), userEmail))

	require.NoError(t, config.RetireCustomField(textID, userEmail))
	require.NoError(t, config.ReactivateCustomField(textID, userEmail))
	require.NoError(t, config.ReorderFields(reverseOrder(config.DisplayOrder()), userEmail))

	events := config.GetUncommittedChanges()
	requireEventTypesPresent(t, events,
		"OnePagerConfigurationCreated",
		"CustomFieldDefined",
		"CustomFieldRenamed",
		"CustomFieldRequirementChanged",
		"CustomFieldRetired",
		"CustomFieldReactivated",
		"BuiltInFieldIncluded",
		"BuiltInFieldExcluded",
		"BuiltInFieldRequirementChanged",
		"OnePagerFieldsReordered",
		"SelectionOptionAdded",
		"SelectionOptionRetired",
		"NumberFieldBoundsChanged",
	)

	loaded := roundTripAndLoad(t, config, len(events))

	assert.Equal(t, config.ID(), loaded.ID())
	assert.Equal(t, config.Version(), loaded.Version())
	assert.Equal(t, config.SubjectType().Value(), loaded.SubjectType().Value())
	assert.Equal(t, config.DisplayOrder(), loaded.DisplayOrder())
	assert.Equal(t, config.CustomFields(), loaded.CustomFields())
}

func TestOnePagerEventDeserializers_CoverEveryConfigurationEventType(t *testing.T) {
	for _, eventType := range events.ConfigurationEventTypes() {
		assert.Truef(t, onePagerEventDeserializers.HasDeserializerFor(eventType),
			"No deserializer registered for %q: the event store silently skips unknown event types, "+
				"so the aggregate reloads a version lower than the stored one and every subsequent "+
				"write fails with a concurrency conflict", eventType)
	}
}

func TestOnePagerFactsEventDeserializers_CoverEveryFactsEventType(t *testing.T) {
	for _, eventType := range events.FactsEventTypes() {
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

func roundTripAndLoad(t *testing.T, config *aggregates.OnePagerConfiguration, expectedEventCount int) *aggregates.OnePagerConfiguration {
	t.Helper()
	events := config.GetUncommittedChanges()
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
