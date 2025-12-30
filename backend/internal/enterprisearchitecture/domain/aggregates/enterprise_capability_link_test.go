package aggregates

import (
	"testing"

	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEnterpriseCapabilityLink(t *testing.T) {
	capability := createTestCapability(t)
	domainCapabilityID := valueobjects.NewDomainCapabilityID()
	linkedBy := valueobjects.MustNewLinkedBy("user@example.com")

	link, err := NewEnterpriseCapabilityLink(capability, domainCapabilityID, linkedBy)
	require.NoError(t, err)

	assert.NotNil(t, link)
	assert.NotEmpty(t, link.ID())
	assert.Equal(t, capability.ID(), link.EnterpriseCapabilityID().Value())
	assert.True(t, domainCapabilityID.Equals(link.DomainCapabilityID()))
	assert.True(t, linkedBy.Equals(link.LinkedBy()))
	assert.False(t, link.LinkedAt().IsZero())
	assert.Len(t, link.GetUncommittedChanges(), 1)
}

func TestNewEnterpriseCapabilityLink_FailsForInactiveCapability(t *testing.T) {
	capability := createTestCapability(t)
	_ = capability.Delete()
	capability.MarkChangesAsCommitted()

	domainCapabilityID := valueobjects.NewDomainCapabilityID()
	linkedBy := valueobjects.MustNewLinkedBy("user@example.com")

	_, err := NewEnterpriseCapabilityLink(capability, domainCapabilityID, linkedBy)
	assert.ErrorIs(t, err, ErrCannotLinkInactiveCapability)
}

func TestEnterpriseCapabilityLink_RaisesCreatedEvent(t *testing.T) {
	capability := createTestCapability(t)
	domainCapabilityID := valueobjects.NewDomainCapabilityID()
	linkedBy := valueobjects.MustNewLinkedBy("user@example.com")

	link, err := NewEnterpriseCapabilityLink(capability, domainCapabilityID, linkedBy)
	require.NoError(t, err)

	uncommittedEvents := link.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "EnterpriseCapabilityLinked", uncommittedEvents[0].EventType())

	eventData := uncommittedEvents[0].EventData()
	assert.Equal(t, link.ID(), eventData["id"])
	assert.Equal(t, capability.ID(), eventData["enterpriseCapabilityId"])
	assert.Equal(t, domainCapabilityID.Value(), eventData["domainCapabilityId"])
	assert.Equal(t, linkedBy.Value(), eventData["linkedBy"])
}

func TestEnterpriseCapabilityLink_Unlink(t *testing.T) {
	capability := createTestCapability(t)
	domainCapabilityID := valueobjects.NewDomainCapabilityID()
	linkedBy := valueobjects.MustNewLinkedBy("user@example.com")

	link, _ := NewEnterpriseCapabilityLink(capability, domainCapabilityID, linkedBy)
	link.MarkChangesAsCommitted()

	err := link.Unlink()
	require.NoError(t, err)

	uncommittedEvents := link.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "EnterpriseCapabilityUnlinked", uncommittedEvents[0].EventType())

	eventData := uncommittedEvents[0].EventData()
	assert.Equal(t, link.ID(), eventData["id"])
	assert.Equal(t, capability.ID(), eventData["enterpriseCapabilityId"])
	assert.Equal(t, domainCapabilityID.Value(), eventData["domainCapabilityId"])
}

func TestEnterpriseCapabilityLink_LoadFromHistory(t *testing.T) {
	capability := createTestCapability(t)
	domainCapabilityID := valueobjects.NewDomainCapabilityID()
	linkedBy := valueobjects.MustNewLinkedBy("user@example.com")

	link, _ := NewEnterpriseCapabilityLink(capability, domainCapabilityID, linkedBy)

	events := link.GetUncommittedChanges()

	loadedLink, err := LoadEnterpriseCapabilityLinkFromHistory(events)
	require.NoError(t, err)

	assert.Equal(t, link.ID(), loadedLink.ID())
	assert.Equal(t, capability.ID(), loadedLink.EnterpriseCapabilityID().Value())
	assert.Equal(t, domainCapabilityID.Value(), loadedLink.DomainCapabilityID().Value())
	assert.Equal(t, linkedBy.Value(), loadedLink.LinkedBy().Value())
}

func TestEnterpriseCapabilityLink_LoadFromHistoryWithUnlink(t *testing.T) {
	capability := createTestCapability(t)
	domainCapabilityID := valueobjects.NewDomainCapabilityID()
	linkedBy := valueobjects.MustNewLinkedBy("user@example.com")

	link, _ := NewEnterpriseCapabilityLink(capability, domainCapabilityID, linkedBy)
	_ = link.Unlink()

	events := link.GetUncommittedChanges()

	loadedLink, err := LoadEnterpriseCapabilityLinkFromHistory(events)
	require.NoError(t, err)

	assert.Equal(t, link.ID(), loadedLink.ID())
}

func TestEnterpriseCapabilityLink_UnlinkMultipleTimes(t *testing.T) {
	capability := createTestCapability(t)
	domainCapabilityID := valueobjects.NewDomainCapabilityID()
	linkedBy := valueobjects.MustNewLinkedBy("user@example.com")

	link, _ := NewEnterpriseCapabilityLink(capability, domainCapabilityID, linkedBy)
	link.MarkChangesAsCommitted()

	err := link.Unlink()
	require.NoError(t, err)

	err = link.Unlink()
	require.NoError(t, err)

	uncommittedEvents := link.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 2)
}

func TestEnterpriseCapabilityLink_LoadFromEmptyHistory(t *testing.T) {
	loadedLink, err := LoadEnterpriseCapabilityLinkFromHistory(nil)
	require.NoError(t, err)

	assert.NotNil(t, loadedLink)
	assert.NotEmpty(t, loadedLink.ID())
	assert.Empty(t, loadedLink.GetUncommittedChanges())
}

func createTestCapability(t *testing.T) *EnterpriseCapability {
	t.Helper()
	name, _ := valueobjects.NewEnterpriseCapabilityName("Test Capability")
	description, _ := valueobjects.NewDescription("Test description")
	category, _ := valueobjects.NewCategory("Differentiating")
	capability, err := NewEnterpriseCapability(name, description, category)
	require.NoError(t, err)
	capability.MarkChangesAsCommitted()
	return capability
}
