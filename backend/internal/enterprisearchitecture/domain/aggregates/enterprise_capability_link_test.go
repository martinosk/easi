package aggregates

import (
	"testing"

	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type linkFixture struct {
	capability         *EnterpriseCapability
	domainCapabilityID valueobjects.DomainCapabilityID
	linkedBy           valueobjects.LinkedBy
}

func setupLinkFixture(t *testing.T) linkFixture {
	t.Helper()
	return linkFixture{
		capability:         createTestCapability(t),
		domainCapabilityID: valueobjects.NewDomainCapabilityID(),
		linkedBy:           valueobjects.MustNewLinkedBy("user@example.com"),
	}
}

func (f linkFixture) newLink(t *testing.T) *EnterpriseCapabilityLink {
	t.Helper()
	link, err := NewEnterpriseCapabilityLink(f.capability, f.domainCapabilityID, f.linkedBy)
	require.NoError(t, err)
	return link
}

func (f linkFixture) assertLinkIdentity(t *testing.T, eventData map[string]interface{}, link *EnterpriseCapabilityLink) {
	t.Helper()
	assert.Equal(t, link.ID(), eventData["id"])
	assert.Equal(t, f.capability.ID(), eventData["enterpriseCapabilityId"])
	assert.Equal(t, f.domainCapabilityID.Value(), eventData["domainCapabilityId"])
}

func TestNewEnterpriseCapabilityLink(t *testing.T) {
	f := setupLinkFixture(t)

	link := f.newLink(t)

	assert.NotNil(t, link)
	assert.NotEmpty(t, link.ID())
	assert.Equal(t, f.capability.ID(), link.EnterpriseCapabilityID().Value())
	assert.True(t, f.domainCapabilityID.Equals(link.DomainCapabilityID()))
	assert.True(t, f.linkedBy.Equals(link.LinkedBy()))
	assert.False(t, link.LinkedAt().IsZero())
	assert.Len(t, link.GetUncommittedChanges(), 1)
}

func TestNewEnterpriseCapabilityLink_FailsForInactiveCapability(t *testing.T) {
	f := setupLinkFixture(t)
	_ = f.capability.Delete()
	f.capability.MarkChangesAsCommitted()

	_, err := NewEnterpriseCapabilityLink(f.capability, f.domainCapabilityID, f.linkedBy)
	assert.ErrorIs(t, err, ErrCannotLinkInactiveCapability)
}

func TestEnterpriseCapabilityLink_RaisesCreatedEvent(t *testing.T) {
	f := setupLinkFixture(t)
	link := f.newLink(t)

	uncommittedEvents := link.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "EnterpriseCapabilityLinked", uncommittedEvents[0].EventType())

	eventData := uncommittedEvents[0].EventData()
	f.assertLinkIdentity(t, eventData, link)
	assert.Equal(t, f.linkedBy.Value(), eventData["linkedBy"])
}

func TestEnterpriseCapabilityLink_Unlink(t *testing.T) {
	f := setupLinkFixture(t)
	link := f.newLink(t)
	link.MarkChangesAsCommitted()

	err := link.Unlink()
	require.NoError(t, err)

	uncommittedEvents := link.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "EnterpriseCapabilityUnlinked", uncommittedEvents[0].EventType())

	f.assertLinkIdentity(t, uncommittedEvents[0].EventData(), link)
}

func TestEnterpriseCapabilityLink_LoadFromHistory(t *testing.T) {
	f := setupLinkFixture(t)
	link := f.newLink(t)

	events := link.GetUncommittedChanges()

	loadedLink, err := LoadEnterpriseCapabilityLinkFromHistory(events)
	require.NoError(t, err)

	assert.Equal(t, link.ID(), loadedLink.ID())
	assert.Equal(t, f.capability.ID(), loadedLink.EnterpriseCapabilityID().Value())
	assert.Equal(t, f.domainCapabilityID.Value(), loadedLink.DomainCapabilityID().Value())
	assert.Equal(t, f.linkedBy.Value(), loadedLink.LinkedBy().Value())
}

func TestEnterpriseCapabilityLink_LoadFromHistoryWithUnlink(t *testing.T) {
	f := setupLinkFixture(t)
	link := f.newLink(t)
	_ = link.Unlink()

	events := link.GetUncommittedChanges()

	loadedLink, err := LoadEnterpriseCapabilityLinkFromHistory(events)
	require.NoError(t, err)

	assert.Equal(t, link.ID(), loadedLink.ID())
}

func TestEnterpriseCapabilityLink_UnlinkMultipleTimes(t *testing.T) {
	f := setupLinkFixture(t)
	link := f.newLink(t)
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
