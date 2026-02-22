package repositories

import (
	"testing"

	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnterpriseCapabilityLinkDeserializers_RoundTrip(t *testing.T) {
	capability := createTestCapability(t)
	domainCapabilityID := valueobjects.NewDomainCapabilityID()
	linkedBy := valueobjects.MustNewLinkedBy("user@example.com")

	link, err := aggregates.NewEnterpriseCapabilityLink(capability, domainCapabilityID, linkedBy)
	require.NoError(t, err)

	events := link.GetUncommittedChanges()
	require.Len(t, events, 1, "Expected 1 event: Linked")

	storedEvents := simulateEventStoreRoundTrip(t, events)
	deserializedEvents, err := enterpriseCapabilityLinkEventDeserializers.Deserialize(storedEvents)
	require.NoError(t, err)

	require.Len(t, deserializedEvents, 1, "All events should be deserialized")

	loaded, err := aggregates.LoadEnterpriseCapabilityLinkFromHistory(deserializedEvents)
	require.NoError(t, err)

	assert.Equal(t, link.ID(), loaded.ID())
	assert.Equal(t, link.EnterpriseCapabilityID().Value(), loaded.EnterpriseCapabilityID().Value())
	assert.Equal(t, link.DomainCapabilityID().Value(), loaded.DomainCapabilityID().Value())
	assert.Equal(t, link.LinkedBy().Value(), loaded.LinkedBy().Value())
}

func TestEnterpriseCapabilityLinkDeserializers_RoundTripWithUnlink(t *testing.T) {
	capability := createTestCapability(t)
	domainCapabilityID := valueobjects.NewDomainCapabilityID()
	linkedBy := valueobjects.MustNewLinkedBy("admin@example.com")

	link, err := aggregates.NewEnterpriseCapabilityLink(capability, domainCapabilityID, linkedBy)
	require.NoError(t, err)

	_ = link.Unlink()

	events := link.GetUncommittedChanges()
	require.Len(t, events, 2, "Expected 2 events: Linked, Unlinked")

	storedEvents := simulateEventStoreRoundTrip(t, events)
	deserializedEvents, err := enterpriseCapabilityLinkEventDeserializers.Deserialize(storedEvents)
	require.NoError(t, err)

	require.Len(t, deserializedEvents, 2, "All events should be deserialized")

	_, err = aggregates.LoadEnterpriseCapabilityLinkFromHistory(deserializedEvents)
	require.NoError(t, err)
}

func TestEnterpriseCapabilityLinkDeserializers_AllEventsCanBeDeserialized(t *testing.T) {
	capability := createTestCapability(t)
	domainCapabilityID := valueobjects.NewDomainCapabilityID()
	linkedBy := valueobjects.MustNewLinkedBy("system")

	link, err := aggregates.NewEnterpriseCapabilityLink(capability, domainCapabilityID, linkedBy)
	require.NoError(t, err)
	_ = link.Unlink()

	events := link.GetUncommittedChanges()

	storedEvents := simulateEventStoreRoundTrip(t, events)
	deserializedEvents, err := enterpriseCapabilityLinkEventDeserializers.Deserialize(storedEvents)
	require.NoError(t, err)

	require.Len(t, deserializedEvents, len(events),
		"All events should be deserialized - missing deserializer for one or more event types")

	for i, original := range events {
		assert.Equal(t, original.EventType(), deserializedEvents[i].EventType(),
			"Event type mismatch at index %d", i)
	}
}
