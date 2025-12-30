package aggregates

import (
	"testing"

	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEnterpriseCapability(t *testing.T) {
	name, err := valueobjects.NewEnterpriseCapabilityName("Payroll")
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Employee compensation processing")
	category, err := valueobjects.NewCategory("HR Operations")
	require.NoError(t, err)

	capability, err := NewEnterpriseCapability(name, description, category)
	require.NoError(t, err)

	assert.NotNil(t, capability)
	assert.NotEmpty(t, capability.ID())
	assert.Equal(t, name, capability.Name())
	assert.Equal(t, description, capability.Description())
	assert.Equal(t, category, capability.Category())
	assert.True(t, capability.IsActive())
	assert.NotZero(t, capability.CreatedAt())
	assert.Len(t, capability.GetUncommittedChanges(), 1)
}

func TestNewEnterpriseCapability_WithEmptyCategory(t *testing.T) {
	name, _ := valueobjects.NewEnterpriseCapabilityName("Customer Identity")
	description := valueobjects.MustNewDescription("Identity management")
	category := valueobjects.EmptyCategory()

	capability, err := NewEnterpriseCapability(name, description, category)
	require.NoError(t, err)

	assert.True(t, capability.Category().IsEmpty())
}

func TestEnterpriseCapability_RaisesCreatedEvent(t *testing.T) {
	name, _ := valueobjects.NewEnterpriseCapabilityName("Payroll")
	description := valueobjects.MustNewDescription("Test description")
	category, _ := valueobjects.NewCategory("HR")

	capability, err := NewEnterpriseCapability(name, description, category)
	require.NoError(t, err)

	uncommittedEvents := capability.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "EnterpriseCapabilityCreated", uncommittedEvents[0].EventType())

	eventData := uncommittedEvents[0].EventData()
	assert.Equal(t, capability.ID(), eventData["id"])
	assert.Equal(t, name.Value(), eventData["name"])
	assert.Equal(t, description.Value(), eventData["description"])
	assert.Equal(t, category.Value(), eventData["category"])
	assert.True(t, eventData["active"].(bool))
}

func TestEnterpriseCapability_Update(t *testing.T) {
	capability := createEnterpriseCapability(t, "Payroll")
	capability.MarkChangesAsCommitted()

	newName, _ := valueobjects.NewEnterpriseCapabilityName("Payroll Management")
	newDescription := valueobjects.MustNewDescription("Updated description")
	newCategory, _ := valueobjects.NewCategory("Finance")

	err := capability.Update(newName, newDescription, newCategory)
	require.NoError(t, err)

	assert.Equal(t, newName, capability.Name())
	assert.Equal(t, newDescription, capability.Description())
	assert.Equal(t, newCategory, capability.Category())

	uncommittedEvents := capability.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "EnterpriseCapabilityUpdated", uncommittedEvents[0].EventType())
}

func TestEnterpriseCapability_Delete_SoftDelete(t *testing.T) {
	capability := createEnterpriseCapability(t, "Payroll")
	capability.MarkChangesAsCommitted()

	assert.True(t, capability.IsActive())

	err := capability.Delete()
	require.NoError(t, err)

	assert.False(t, capability.IsActive())

	uncommittedEvents := capability.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "EnterpriseCapabilityDeleted", uncommittedEvents[0].EventType())
}

func TestEnterpriseCapability_LoadFromHistory(t *testing.T) {
	name, _ := valueobjects.NewEnterpriseCapabilityName("Payroll")
	description := valueobjects.MustNewDescription("Test description")
	category, _ := valueobjects.NewCategory("HR")

	capability, _ := NewEnterpriseCapability(name, description, category)

	events := capability.GetUncommittedChanges()

	loadedCapability, err := LoadEnterpriseCapabilityFromHistory(events)
	require.NoError(t, err)

	assert.Equal(t, capability.ID(), loadedCapability.ID())
	assert.Equal(t, capability.Name().Value(), loadedCapability.Name().Value())
	assert.Equal(t, capability.Description().Value(), loadedCapability.Description().Value())
	assert.Equal(t, capability.Category().Value(), loadedCapability.Category().Value())
	assert.Equal(t, capability.IsActive(), loadedCapability.IsActive())
}

func TestEnterpriseCapability_LoadFromHistoryWithUpdate(t *testing.T) {
	capability := createEnterpriseCapability(t, "Payroll")

	newName, _ := valueobjects.NewEnterpriseCapabilityName("Payroll Management")
	newDescription := valueobjects.MustNewDescription("Updated")
	newCategory, _ := valueobjects.NewCategory("Finance")

	_ = capability.Update(newName, newDescription, newCategory)

	events := capability.GetUncommittedChanges()

	loadedCapability, err := LoadEnterpriseCapabilityFromHistory(events)
	require.NoError(t, err)

	assert.Equal(t, newName.Value(), loadedCapability.Name().Value())
	assert.Equal(t, newDescription.Value(), loadedCapability.Description().Value())
	assert.Equal(t, newCategory.Value(), loadedCapability.Category().Value())
}

func TestEnterpriseCapability_LoadFromHistoryWithDelete(t *testing.T) {
	capability := createEnterpriseCapability(t, "Payroll")
	_ = capability.Delete()

	events := capability.GetUncommittedChanges()

	loadedCapability, err := LoadEnterpriseCapabilityFromHistory(events)
	require.NoError(t, err)

	assert.False(t, loadedCapability.IsActive())
}

func TestEnterpriseCapability_UpdateAfterDelete(t *testing.T) {
	capability := createEnterpriseCapability(t, "Payroll")
	capability.MarkChangesAsCommitted()

	err := capability.Delete()
	require.NoError(t, err)
	capability.MarkChangesAsCommitted()

	assert.False(t, capability.IsActive())

	newName, _ := valueobjects.NewEnterpriseCapabilityName("Updated Payroll")
	newDescription := valueobjects.MustNewDescription("Updated after delete")
	newCategory, _ := valueobjects.NewCategory("Updated")

	err = capability.Update(newName, newDescription, newCategory)
	require.NoError(t, err)

	assert.Equal(t, newName, capability.Name())
	assert.Equal(t, newDescription, capability.Description())
	assert.Equal(t, newCategory, capability.Category())
}

func TestEnterpriseCapability_LoadFromEmptyHistory(t *testing.T) {
	loadedCapability, err := LoadEnterpriseCapabilityFromHistory(nil)
	require.NoError(t, err)

	assert.NotNil(t, loadedCapability)
	assert.NotEmpty(t, loadedCapability.ID())
	assert.Empty(t, loadedCapability.GetUncommittedChanges())
}

func TestEnterpriseCapability_MultipleUpdates(t *testing.T) {
	capability := createEnterpriseCapability(t, "Original")
	capability.MarkChangesAsCommitted()

	for i := 0; i < 3; i++ {
		newName, _ := valueobjects.NewEnterpriseCapabilityName("Updated " + string(rune('A'+i)))
		newDescription := valueobjects.MustNewDescription("Description " + string(rune('A'+i)))
		newCategory, _ := valueobjects.NewCategory("Category")

		err := capability.Update(newName, newDescription, newCategory)
		require.NoError(t, err)
	}

	uncommittedEvents := capability.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 3)

	assert.Equal(t, "Updated C", capability.Name().Value())
}

func createEnterpriseCapability(t *testing.T, capabilityName string) *EnterpriseCapability {
	t.Helper()

	name, err := valueobjects.NewEnterpriseCapabilityName(capabilityName)
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Test enterprise capability")
	category, _ := valueobjects.NewCategory("Test Category")

	capability, err := NewEnterpriseCapability(name, description, category)
	require.NoError(t, err)

	return capability
}
