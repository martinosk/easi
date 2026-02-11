package aggregates

import (
	"testing"

	"easi/backend/internal/valuestreams/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValueStream(t *testing.T) {
	name, err := valueobjects.NewValueStreamName("Customer Onboarding")
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("End-to-end customer onboarding process")

	vs, err := NewValueStream(name, description)
	require.NoError(t, err)
	assert.NotNil(t, vs)
	assert.NotEmpty(t, vs.ID())
	assert.Equal(t, name, vs.Name())
	assert.Equal(t, description, vs.Description())
	assert.NotZero(t, vs.CreatedAt())
	assert.Len(t, vs.GetUncommittedChanges(), 1)
}

func TestValueStream_RaisesCreatedEvent(t *testing.T) {
	name, err := valueobjects.NewValueStreamName("Order Fulfillment")
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Order processing and delivery")

	vs, err := NewValueStream(name, description)
	require.NoError(t, err)

	uncommittedEvents := vs.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "ValueStreamCreated", uncommittedEvents[0].EventType())

	eventData := uncommittedEvents[0].EventData()
	assert.Equal(t, vs.ID(), eventData["id"])
	assert.Equal(t, name.Value(), eventData["name"])
	assert.Equal(t, description.Value(), eventData["description"])
	assert.NotNil(t, eventData["createdAt"])
}

func TestValueStream_Update(t *testing.T) {
	vs := createValueStream(t, "Customer Onboarding")
	vs.MarkChangesAsCommitted()

	newName, err := valueobjects.NewValueStreamName("Customer Onboarding v2")
	require.NoError(t, err)

	newDescription := valueobjects.MustNewDescription("Updated onboarding process")

	err = vs.Update(newName, newDescription)
	require.NoError(t, err)

	assert.Equal(t, newName, vs.Name())
	assert.Equal(t, newDescription, vs.Description())

	uncommittedEvents := vs.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "ValueStreamUpdated", uncommittedEvents[0].EventType())
}

func TestValueStream_UpdateRaisesEvent(t *testing.T) {
	vs := createValueStream(t, "Order Fulfillment")
	vs.MarkChangesAsCommitted()

	newName, err := valueobjects.NewValueStreamName("Order Fulfillment v2")
	require.NoError(t, err)

	newDescription := valueobjects.MustNewDescription("Updated fulfillment")

	err = vs.Update(newName, newDescription)
	require.NoError(t, err)

	uncommittedEvents := vs.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)

	event := uncommittedEvents[0]
	assert.Equal(t, "ValueStreamUpdated", event.EventType())

	eventData := event.EventData()
	assert.Equal(t, vs.ID(), eventData["id"])
	assert.Equal(t, newName.Value(), eventData["name"])
	assert.Equal(t, newDescription.Value(), eventData["description"])
	assert.NotNil(t, eventData["updatedAt"])
}

func TestValueStream_Delete(t *testing.T) {
	vs := createValueStream(t, "Customer Onboarding")
	vs.MarkChangesAsCommitted()

	err := vs.Delete()
	require.NoError(t, err)

	uncommittedEvents := vs.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "ValueStreamDeleted", uncommittedEvents[0].EventType())

	eventData := uncommittedEvents[0].EventData()
	assert.Equal(t, vs.ID(), eventData["id"])
	assert.NotNil(t, eventData["deletedAt"])
}

func TestValueStream_DeletePreservesState(t *testing.T) {
	name, err := valueobjects.NewValueStreamName("Customer Onboarding")
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Onboarding process")

	vs, err := NewValueStream(name, description)
	require.NoError(t, err)
	vs.MarkChangesAsCommitted()

	originalID := vs.ID()
	originalName := vs.Name().Value()

	err = vs.Delete()
	require.NoError(t, err)

	assert.Equal(t, originalID, vs.ID())
	assert.Equal(t, originalName, vs.Name().Value())
}

func TestValueStream_LoadFromHistory(t *testing.T) {
	name, err := valueobjects.NewValueStreamName("Order Fulfillment")
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Fulfillment process")

	vs, err := NewValueStream(name, description)
	require.NoError(t, err)

	events := vs.GetUncommittedChanges()

	loadedVS, err := LoadValueStreamFromHistory(events)
	require.NoError(t, err)
	assert.NotNil(t, loadedVS)
	assert.Equal(t, vs.ID(), loadedVS.ID())
	assert.Equal(t, vs.Name().Value(), loadedVS.Name().Value())
	assert.Equal(t, vs.Description().Value(), loadedVS.Description().Value())
}

func TestValueStream_LoadFromHistoryWithMultipleEvents(t *testing.T) {
	name, err := valueobjects.NewValueStreamName("Customer Onboarding")
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Onboarding process")

	vs, err := NewValueStream(name, description)
	require.NoError(t, err)

	newName, err := valueobjects.NewValueStreamName("Customer Onboarding v2")
	require.NoError(t, err)

	newDescription := valueobjects.MustNewDescription("Updated onboarding")

	err = vs.Update(newName, newDescription)
	require.NoError(t, err)

	allEvents := vs.GetUncommittedChanges()

	loadedVS, err := LoadValueStreamFromHistory(allEvents)
	require.NoError(t, err)

	assert.Equal(t, vs.ID(), loadedVS.ID())
	assert.Equal(t, newName.Value(), loadedVS.Name().Value())
	assert.Equal(t, newDescription.Value(), loadedVS.Description().Value())
}

func TestValueStream_LoadFromHistoryWithDelete(t *testing.T) {
	vs := createValueStream(t, "Customer Onboarding")

	err := vs.Delete()
	require.NoError(t, err)

	allEvents := vs.GetUncommittedChanges()
	require.Len(t, allEvents, 2)

	loadedVS, err := LoadValueStreamFromHistory(allEvents)
	require.NoError(t, err)
	assert.Equal(t, vs.ID(), loadedVS.ID())
	assert.Equal(t, vs.Name().Value(), loadedVS.Name().Value())
}

func createValueStream(t *testing.T, vsName string) *ValueStream {
	t.Helper()

	name, err := valueobjects.NewValueStreamName(vsName)
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Test value stream")

	vs, err := NewValueStream(name, description)
	require.NoError(t, err)

	return vs
}
