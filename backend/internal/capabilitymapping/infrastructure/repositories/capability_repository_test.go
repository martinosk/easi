package repositories

import (
	"encoding/json"
	"testing"
	"time"

	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/entities"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCapabilityDeserializers_RoundTrip(t *testing.T) {
	name, _ := valueobjects.NewCapabilityName("Order Management")
	description := valueobjects.MustNewDescription("Manages customer orders")

	original, err := aggregates.NewCapability(name, description, valueobjects.CapabilityID{}, valueobjects.LevelL1)
	require.NoError(t, err)

	events := original.GetUncommittedChanges()
	require.Len(t, events, 1, "Expected 1 event: Created")

	storedEvents := simulateCapabilityEventStoreRoundTrip(t, events)
	deserializedEvents := capabilityEventDeserializers.Deserialize(storedEvents)

	require.Len(t, deserializedEvents, 1, "All events should be deserialized")

	loaded, err := aggregates.LoadCapabilityFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.Equal(t, original.ID(), loaded.ID())
	assert.Equal(t, original.Name().Value(), loaded.Name().Value())
	assert.Equal(t, original.Description().Value(), loaded.Description().Value())
	assert.Equal(t, original.Level().Value(), loaded.Level().Value())
}

func TestCapabilityDeserializers_RoundTripWithUpdate(t *testing.T) {
	name, _ := valueobjects.NewCapabilityName("Payment Processing")
	description := valueobjects.MustNewDescription("Handles payments")

	original, err := aggregates.NewCapability(name, description, valueobjects.CapabilityID{}, valueobjects.LevelL1)
	require.NoError(t, err)

	newName, _ := valueobjects.NewCapabilityName("Payment Gateway")
	newDescription := valueobjects.MustNewDescription("Updated payment description")
	_ = original.Update(newName, newDescription)

	events := original.GetUncommittedChanges()
	require.Len(t, events, 2, "Expected 2 events: Created, Updated")

	storedEvents := simulateCapabilityEventStoreRoundTrip(t, events)
	deserializedEvents := capabilityEventDeserializers.Deserialize(storedEvents)

	require.Len(t, deserializedEvents, 2, "All events should be deserialized")

	loaded, err := aggregates.LoadCapabilityFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.Equal(t, newName.Value(), loaded.Name().Value())
	assert.Equal(t, newDescription.Value(), loaded.Description().Value())
}

func TestCapabilityDeserializers_RoundTripWithMetadata(t *testing.T) {
	name, _ := valueobjects.NewCapabilityName("Inventory Management")
	description := valueobjects.MustNewDescription("Manages inventory")

	original, err := aggregates.NewCapability(name, description, valueobjects.CapabilityID{}, valueobjects.LevelL1)
	require.NoError(t, err)

	maturityLevel, _ := valueobjects.NewMaturityLevelFromValue(50)
	ownershipModel, _ := valueobjects.NewOwnershipModel("IT")
	primaryOwner := valueobjects.NewOwner("john@example.com")
	eaOwner := valueobjects.NewOwner("ea@example.com")
	status, _ := valueobjects.NewCapabilityStatus("active")

	metadata := valueobjects.NewCapabilityMetadata(maturityLevel, ownershipModel, primaryOwner, eaOwner, status)
	_ = original.UpdateMetadata(metadata)

	events := original.GetUncommittedChanges()
	require.Len(t, events, 2, "Expected 2 events: Created, MetadataUpdated")

	storedEvents := simulateCapabilityEventStoreRoundTrip(t, events)
	deserializedEvents := capabilityEventDeserializers.Deserialize(storedEvents)

	require.Len(t, deserializedEvents, 2, "All events should be deserialized")

	loaded, err := aggregates.LoadCapabilityFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.Equal(t, maturityLevel.Value(), loaded.MaturityLevel().Value())
	assert.Equal(t, ownershipModel.Value(), loaded.OwnershipModel().Value())
}

func TestCapabilityDeserializers_RoundTripWithExpertAndTag(t *testing.T) {
	name, _ := valueobjects.NewCapabilityName("Customer Service")
	description := valueobjects.MustNewDescription("Handles customer inquiries")

	original, err := aggregates.NewCapability(name, description, valueobjects.CapabilityID{}, valueobjects.LevelL1)
	require.NoError(t, err)

	expert, _ := entities.NewExpert("Jane Doe", "Domain Expert", "jane@example.com")
	_ = original.AddExpert(expert)

	tag, _ := valueobjects.NewTag("core-capability")
	_ = original.AddTag(tag)

	events := original.GetUncommittedChanges()
	require.Len(t, events, 3, "Expected 3 events: Created, ExpertAdded, TagAdded")

	storedEvents := simulateCapabilityEventStoreRoundTrip(t, events)
	deserializedEvents := capabilityEventDeserializers.Deserialize(storedEvents)

	require.Len(t, deserializedEvents, 3, "All events should be deserialized")

	loaded, err := aggregates.LoadCapabilityFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.Len(t, loaded.Experts(), 1)
	assert.Len(t, loaded.Tags(), 1)
}

func TestCapabilityDeserializers_RoundTripWithParentChange(t *testing.T) {
	name, _ := valueobjects.NewCapabilityName("Order Fulfillment")
	description := valueobjects.MustNewDescription("Fulfills orders")

	original, err := aggregates.NewCapability(name, description, valueobjects.CapabilityID{}, valueobjects.LevelL1)
	require.NoError(t, err)

	parentCapability, _ := aggregates.NewCapability(name, description, valueobjects.CapabilityID{}, valueobjects.LevelL1)
	newParentID, _ := valueobjects.NewCapabilityIDFromString(parentCapability.ID())
	_ = original.ChangeParent(newParentID, valueobjects.LevelL2)

	events := original.GetUncommittedChanges()
	require.Len(t, events, 2, "Expected 2 events: Created, ParentChanged")

	storedEvents := simulateCapabilityEventStoreRoundTrip(t, events)
	deserializedEvents := capabilityEventDeserializers.Deserialize(storedEvents)

	require.Len(t, deserializedEvents, 2, "All events should be deserialized")

	loaded, err := aggregates.LoadCapabilityFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.Equal(t, valueobjects.LevelL2.Value(), loaded.Level().Value())
}

func TestCapabilityDeserializers_AllEventsCanBeDeserialized(t *testing.T) {
	name, _ := valueobjects.NewCapabilityName("Test Capability")
	description := valueobjects.MustNewDescription("Test description")

	capability, err := aggregates.NewCapability(name, description, valueobjects.CapabilityID{}, valueobjects.LevelL1)
	require.NoError(t, err)

	newName, _ := valueobjects.NewCapabilityName("Updated Name")
	newDescription := valueobjects.MustNewDescription("Updated description")
	_ = capability.Update(newName, newDescription)

	maturityLevel, _ := valueobjects.NewMaturityLevelFromValue(30)
	ownershipModel, _ := valueobjects.NewOwnershipModel("Business")
	primaryOwner := valueobjects.NewOwner("owner@example.com")
	eaOwner := valueobjects.NewOwner("ea@example.com")
	status, _ := valueobjects.NewCapabilityStatus("active")
	metadata := valueobjects.NewCapabilityMetadata(maturityLevel, ownershipModel, primaryOwner, eaOwner, status)
	_ = capability.UpdateMetadata(metadata)

	expert, _ := entities.NewExpert("Expert", "Role", "contact")
	_ = capability.AddExpert(expert)

	tag, _ := valueobjects.NewTag("test-tag")
	_ = capability.AddTag(tag)

	parentCapability, _ := aggregates.NewCapability(name, description, valueobjects.CapabilityID{}, valueobjects.LevelL1)
	newParentID, _ := valueobjects.NewCapabilityIDFromString(parentCapability.ID())
	_ = capability.ChangeParent(newParentID, valueobjects.LevelL2)

	events := capability.GetUncommittedChanges()
	require.Len(t, events, 6, "Expected 6 events")

	storedEvents := simulateCapabilityEventStoreRoundTrip(t, events)
	deserializedEvents := capabilityEventDeserializers.Deserialize(storedEvents)

	require.Len(t, deserializedEvents, len(events),
		"All events should be deserialized - missing deserializer for one or more event types")

	for i, originalEvent := range events {
		assert.Equal(t, originalEvent.EventType(), deserializedEvents[i].EventType(),
			"Event type mismatch at index %d", i)
	}
}

type capabilityStoredEventWrapper struct {
	eventType string
	eventData map[string]interface{}
}

func (e *capabilityStoredEventWrapper) EventType() string                 { return e.eventType }
func (e *capabilityStoredEventWrapper) EventData() map[string]interface{} { return e.eventData }
func (e *capabilityStoredEventWrapper) AggregateID() string               { return "" }
func (e *capabilityStoredEventWrapper) OccurredAt() time.Time             { return time.Time{} }

func simulateCapabilityEventStoreRoundTrip(t *testing.T, events []domain.DomainEvent) []domain.DomainEvent {
	t.Helper()

	result := make([]domain.DomainEvent, len(events))

	for i, event := range events {
		jsonBytes, err := json.Marshal(event.EventData())
		require.NoError(t, err, "Failed to serialize event: %s", event.EventType())

		var data map[string]interface{}
		err = json.Unmarshal(jsonBytes, &data)
		require.NoError(t, err, "Failed to unmarshal JSON for event: %s", event.EventType())

		result[i] = &capabilityStoredEventWrapper{
			eventType: event.EventType(),
			eventData: data,
		}
	}

	return result
}
