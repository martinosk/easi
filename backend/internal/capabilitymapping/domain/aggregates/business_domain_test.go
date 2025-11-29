package aggregates

import (
	"testing"

	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBusinessDomain(t *testing.T) {
	name, err := valueobjects.NewDomainName("Finance")
	require.NoError(t, err)

	description := valueobjects.NewDescription("Financial business domain")

	domain, err := NewBusinessDomain(name, description)
	require.NoError(t, err)
	assert.NotNil(t, domain)
	assert.NotEmpty(t, domain.ID())
	assert.Equal(t, name, domain.Name())
	assert.Equal(t, description, domain.Description())
	assert.NotZero(t, domain.CreatedAt())
	assert.Len(t, domain.GetUncommittedChanges(), 1)
}

func TestBusinessDomain_RaisesCreatedEvent(t *testing.T) {
	name, err := valueobjects.NewDomainName("Customer Experience")
	require.NoError(t, err)

	description := valueobjects.NewDescription("Customer-facing business domain")

	domain, err := NewBusinessDomain(name, description)
	require.NoError(t, err)

	uncommittedEvents := domain.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "BusinessDomainCreated", uncommittedEvents[0].EventType())

	eventData := uncommittedEvents[0].EventData()
	assert.Equal(t, domain.ID(), eventData["id"])
	assert.Equal(t, name.Value(), eventData["name"])
	assert.Equal(t, description.Value(), eventData["description"])
	assert.NotNil(t, eventData["createdAt"])
}

func TestBusinessDomain_Update(t *testing.T) {
	name, err := valueobjects.NewDomainName("Finance")
	require.NoError(t, err)

	description := valueobjects.NewDescription("Financial capabilities")

	domain, err := NewBusinessDomain(name, description)
	require.NoError(t, err)

	domain.MarkChangesAsCommitted()

	newName, err := valueobjects.NewDomainName("Finance & Accounting")
	require.NoError(t, err)

	newDescription := valueobjects.NewDescription("Financial and accounting capabilities")

	err = domain.Update(newName, newDescription)
	require.NoError(t, err)

	assert.Equal(t, newName, domain.Name())
	assert.Equal(t, newDescription, domain.Description())

	uncommittedEvents := domain.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "BusinessDomainUpdated", uncommittedEvents[0].EventType())
}

func TestBusinessDomain_UpdateRaisesEvent(t *testing.T) {
	name, err := valueobjects.NewDomainName("Operations")
	require.NoError(t, err)

	description := valueobjects.NewDescription("Operational domain")

	domain, err := NewBusinessDomain(name, description)
	require.NoError(t, err)

	domain.MarkChangesAsCommitted()

	newName, err := valueobjects.NewDomainName("Operations & Support")
	require.NoError(t, err)

	newDescription := valueobjects.NewDescription("Operations and support domain")

	err = domain.Update(newName, newDescription)
	require.NoError(t, err)

	uncommittedEvents := domain.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)

	event := uncommittedEvents[0]
	assert.Equal(t, "BusinessDomainUpdated", event.EventType())

	eventData := event.EventData()
	assert.Equal(t, domain.ID(), eventData["id"])
	assert.Equal(t, newName.Value(), eventData["name"])
	assert.Equal(t, newDescription.Value(), eventData["description"])
	assert.NotNil(t, eventData["updatedAt"])
}

func TestBusinessDomain_Delete(t *testing.T) {
	domain := createBusinessDomain(t, "Finance")
	domain.MarkChangesAsCommitted()

	err := domain.Delete()
	require.NoError(t, err)

	uncommittedEvents := domain.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "BusinessDomainDeleted", uncommittedEvents[0].EventType())

	eventData := uncommittedEvents[0].EventData()
	assert.Equal(t, domain.ID(), eventData["id"])
	assert.NotNil(t, eventData["deletedAt"])
}

func TestBusinessDomain_DeletePreservesState(t *testing.T) {
	name, err := valueobjects.NewDomainName("Finance")
	require.NoError(t, err)

	description := valueobjects.NewDescription("Financial capabilities")

	domain, err := NewBusinessDomain(name, description)
	require.NoError(t, err)
	domain.MarkChangesAsCommitted()

	originalID := domain.ID()
	originalName := domain.Name().Value()

	err = domain.Delete()
	require.NoError(t, err)

	assert.Equal(t, originalID, domain.ID())
	assert.Equal(t, originalName, domain.Name().Value())
}

func TestBusinessDomain_LoadFromHistory(t *testing.T) {
	name, err := valueobjects.NewDomainName("Operations")
	require.NoError(t, err)

	description := valueobjects.NewDescription("Operational capabilities")

	domain, err := NewBusinessDomain(name, description)
	require.NoError(t, err)

	events := domain.GetUncommittedChanges()

	loadedDomain, err := LoadBusinessDomainFromHistory(events)
	require.NoError(t, err)
	assert.NotNil(t, loadedDomain)
	assert.Equal(t, domain.ID(), loadedDomain.ID())
	assert.Equal(t, domain.Name().Value(), loadedDomain.Name().Value())
	assert.Equal(t, domain.Description().Value(), loadedDomain.Description().Value())
}

func TestBusinessDomain_LoadFromHistoryWithMultipleEvents(t *testing.T) {
	name, err := valueobjects.NewDomainName("Finance")
	require.NoError(t, err)

	description := valueobjects.NewDescription("Financial domain")

	domain, err := NewBusinessDomain(name, description)
	require.NoError(t, err)

	newName, err := valueobjects.NewDomainName("Finance & Accounting")
	require.NoError(t, err)

	newDescription := valueobjects.NewDescription("Financial and accounting domain")

	err = domain.Update(newName, newDescription)
	require.NoError(t, err)

	allEvents := domain.GetUncommittedChanges()

	loadedDomain, err := LoadBusinessDomainFromHistory(allEvents)
	require.NoError(t, err)

	assert.Equal(t, domain.ID(), loadedDomain.ID())
	assert.Equal(t, newName.Value(), loadedDomain.Name().Value())
	assert.Equal(t, newDescription.Value(), loadedDomain.Description().Value())
}

func TestBusinessDomain_LoadFromHistoryWithDelete(t *testing.T) {
	domain := createBusinessDomain(t, "Finance")

	err := domain.Delete()
	require.NoError(t, err)

	allEvents := domain.GetUncommittedChanges()
	require.Len(t, allEvents, 2)

	loadedDomain, err := LoadBusinessDomainFromHistory(allEvents)
	require.NoError(t, err)
	assert.Equal(t, domain.ID(), loadedDomain.ID())
	assert.Equal(t, domain.Name().Value(), loadedDomain.Name().Value())
}

func createBusinessDomain(t *testing.T, domainName string) *BusinessDomain {
	t.Helper()

	name, err := valueobjects.NewDomainName(domainName)
	require.NoError(t, err)

	description := valueobjects.NewDescription("Test business domain")

	domain, err := NewBusinessDomain(name, description)
	require.NoError(t, err)

	return domain
}
