package aggregates

import (
	"testing"

	"easi/backend/internal/architectureviews/domain/valueobjects"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewArchitectureView_ValidInputs(t *testing.T) {
	name, err := valueobjects.NewViewName("Main View")
	require.NoError(t, err)

	description := "Primary architecture view"

	view, err := NewArchitectureView(name, description, false)

	require.NoError(t, err)
	assert.NotNil(t, view)
	assert.NotEmpty(t, view.ID())
	assert.Equal(t, name, view.Name())
	assert.Equal(t, description, view.Description())
	assert.False(t, view.IsDefault())
	assert.False(t, view.IsDeleted())
	assert.NotZero(t, view.CreatedAt())
	assert.Empty(t, view.Components())
}

func TestNewArchitectureView_AsDefault(t *testing.T) {
	name, err := valueobjects.NewViewName("Default View")
	require.NoError(t, err)

	view, err := NewArchitectureView(name, "Default view", true)

	require.NoError(t, err)
	assert.True(t, view.IsDefault())

	// Should have both ViewCreated and DefaultViewChanged events
	events := view.GetUncommittedChanges()
	assert.Len(t, events, 2)
	assert.Equal(t, "ViewCreated", events[0].EventType())
	assert.Equal(t, "DefaultViewChanged", events[1].EventType())
}

func TestNewArchitectureView_RaisesCreatedEvent(t *testing.T) {
	name, err := valueobjects.NewViewName("Test View")
	require.NoError(t, err)

	view, err := NewArchitectureView(name, "Test description", false)
	require.NoError(t, err)

	uncommittedEvents := view.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "ViewCreated", uncommittedEvents[0].EventType())
}

func TestArchitectureView_AddComponent(t *testing.T) {
	name, err := valueobjects.NewViewName("Test View")
	require.NoError(t, err)

	view, err := NewArchitectureView(name, "Test view", false)
	require.NoError(t, err)

	componentID := uuid.New().String()
	position := valueobjects.NewComponentPosition(100, 200)

	err = view.AddComponent(componentID, position)

	require.NoError(t, err)
	components := view.Components()
	assert.Len(t, components, 1)
	assert.Contains(t, components, componentID)
}

func TestArchitectureView_AddComponent_AlreadyExists(t *testing.T) {
	name, err := valueobjects.NewViewName("Test View")
	require.NoError(t, err)

	view, err := NewArchitectureView(name, "Test view", false)
	require.NoError(t, err)

	componentID := uuid.New().String()
	position := valueobjects.NewComponentPosition(100, 200)

	// Add component first time
	err = view.AddComponent(componentID, position)
	require.NoError(t, err)

	// Try to add same component again
	err = view.AddComponent(componentID, position)

	assert.Error(t, err)
	assert.Equal(t, ErrComponentAlreadyInView, err)
}

func TestArchitectureView_UpdateComponentPosition(t *testing.T) {
	name, err := valueobjects.NewViewName("Test View")
	require.NoError(t, err)

	view, err := NewArchitectureView(name, "Test view", false)
	require.NoError(t, err)

	componentID := uuid.New().String()
	initialPosition := valueobjects.NewComponentPosition(100, 200)

	err = view.AddComponent(componentID, initialPosition)
	require.NoError(t, err)

	// Update position
	newPosition := valueobjects.NewComponentPosition(300, 400)
	err = view.UpdateComponentPosition(componentID, newPosition)

	require.NoError(t, err)
	components := view.Components()
	updatedComponent := components[componentID]
	assert.Equal(t, 300.0, updatedComponent.Position().X())
	assert.Equal(t, 400.0, updatedComponent.Position().Y())
}

func TestArchitectureView_UpdateComponentPosition_NotFound(t *testing.T) {
	name, err := valueobjects.NewViewName("Test View")
	require.NoError(t, err)

	view, err := NewArchitectureView(name, "Test view", false)
	require.NoError(t, err)

	nonExistentComponentID := uuid.New().String()
	position := valueobjects.NewComponentPosition(100, 200)

	err = view.UpdateComponentPosition(nonExistentComponentID, position)

	assert.Error(t, err)
	assert.Equal(t, ErrComponentNotFound, err)
}

func TestArchitectureView_RemoveComponent(t *testing.T) {
	name, err := valueobjects.NewViewName("Test View")
	require.NoError(t, err)

	view, err := NewArchitectureView(name, "Test view", false)
	require.NoError(t, err)

	componentID := uuid.New().String()
	position := valueobjects.NewComponentPosition(100, 200)

	err = view.AddComponent(componentID, position)
	require.NoError(t, err)

	// Remove component
	err = view.RemoveComponent(componentID)

	require.NoError(t, err)
	components := view.Components()
	assert.Len(t, components, 0)
	assert.NotContains(t, components, componentID)
}

func TestArchitectureView_RemoveComponent_NotFound(t *testing.T) {
	name, err := valueobjects.NewViewName("Test View")
	require.NoError(t, err)

	view, err := NewArchitectureView(name, "Test view", false)
	require.NoError(t, err)

	nonExistentComponentID := uuid.New().String()

	err = view.RemoveComponent(nonExistentComponentID)

	assert.Error(t, err)
	assert.Equal(t, ErrComponentNotFound, err)
}

func TestArchitectureView_Rename(t *testing.T) {
	name, err := valueobjects.NewViewName("Original Name")
	require.NoError(t, err)

	view, err := NewArchitectureView(name, "Test view", false)
	require.NoError(t, err)

	view.MarkChangesAsCommitted()

	newName, err := valueobjects.NewViewName("New Name")
	require.NoError(t, err)

	err = view.Rename(newName)

	require.NoError(t, err)
	assert.Equal(t, newName, view.Name())

	// Verify rename event was raised
	events := view.GetUncommittedChanges()
	assert.Len(t, events, 1)
	assert.Equal(t, "ViewRenamed", events[0].EventType())
}

func TestArchitectureView_Rename_SameName(t *testing.T) {
	name, err := valueobjects.NewViewName("Same Name")
	require.NoError(t, err)

	view, err := NewArchitectureView(name, "Test view", false)
	require.NoError(t, err)

	view.MarkChangesAsCommitted()

	// Try to rename to the same name
	err = view.Rename(name)

	require.NoError(t, err)
	// No event should be raised
	events := view.GetUncommittedChanges()
	assert.Len(t, events, 0)
}

func TestArchitectureView_Delete(t *testing.T) {
	name, err := valueobjects.NewViewName("Test View")
	require.NoError(t, err)

	view, err := NewArchitectureView(name, "Test view", false)
	require.NoError(t, err)

	err = view.Delete()

	require.NoError(t, err)
	assert.True(t, view.IsDeleted())

	// Verify delete event was raised
	events := view.GetUncommittedChanges()
	assert.Equal(t, "ViewDeleted", events[len(events)-1].EventType())
}

func TestArchitectureView_CannotDeleteDefaultView(t *testing.T) {
	name, err := valueobjects.NewViewName("Default View")
	require.NoError(t, err)

	view, err := NewArchitectureView(name, "Default view", true)
	require.NoError(t, err)

	err = view.Delete()

	assert.Error(t, err)
	assert.Equal(t, ErrCannotDeleteDefaultView, err)
	assert.False(t, view.IsDeleted())
}

func TestArchitectureView_CannotDeleteAlreadyDeletedView(t *testing.T) {
	name, err := valueobjects.NewViewName("Test View")
	require.NoError(t, err)

	view, err := NewArchitectureView(name, "Test view", false)
	require.NoError(t, err)

	// Delete once
	err = view.Delete()
	require.NoError(t, err)

	// Try to delete again
	err = view.Delete()

	assert.Error(t, err)
	assert.Equal(t, ErrViewAlreadyDeleted, err)
}

func TestArchitectureView_SetAsDefault(t *testing.T) {
	name, err := valueobjects.NewViewName("Test View")
	require.NoError(t, err)

	view, err := NewArchitectureView(name, "Test view", false)
	require.NoError(t, err)

	assert.False(t, view.IsDefault())

	err = view.SetAsDefault()

	require.NoError(t, err)
	assert.True(t, view.IsDefault())

	// Verify default changed event was raised
	events := view.GetUncommittedChanges()
	assert.Equal(t, "DefaultViewChanged", events[len(events)-1].EventType())
}

func TestArchitectureView_SetAsDefault_AlreadyDefault(t *testing.T) {
	name, err := valueobjects.NewViewName("Default View")
	require.NoError(t, err)

	view, err := NewArchitectureView(name, "Default view", true)
	require.NoError(t, err)

	view.MarkChangesAsCommitted()

	// Try to set as default when already default
	err = view.SetAsDefault()

	require.NoError(t, err)
	// No new event should be raised
	events := view.GetUncommittedChanges()
	assert.Len(t, events, 0)
}

func TestArchitectureView_UnsetAsDefault(t *testing.T) {
	name, err := valueobjects.NewViewName("Default View")
	require.NoError(t, err)

	view, err := NewArchitectureView(name, "Default view", true)
	require.NoError(t, err)

	assert.True(t, view.IsDefault())

	err = view.UnsetAsDefault()

	require.NoError(t, err)
	assert.False(t, view.IsDefault())

	// Verify default changed event was raised
	events := view.GetUncommittedChanges()
	assert.Equal(t, "DefaultViewChanged", events[len(events)-1].EventType())
}

func TestArchitectureView_CannotOperateOnDeletedView(t *testing.T) {
	name, err := valueobjects.NewViewName("Test View")
	require.NoError(t, err)

	view, err := NewArchitectureView(name, "Test view", false)
	require.NoError(t, err)

	// Delete the view
	err = view.Delete()
	require.NoError(t, err)

	// Try various operations on deleted view
	componentID := uuid.New().String()
	position := valueobjects.NewComponentPosition(100, 200)

	err = view.AddComponent(componentID, position)
	assert.Equal(t, ErrViewAlreadyDeleted, err)

	err = view.UpdateComponentPosition(componentID, position)
	assert.Equal(t, ErrViewAlreadyDeleted, err)

	err = view.RemoveComponent(componentID)
	assert.Equal(t, ErrViewAlreadyDeleted, err)

	newName, _ := valueobjects.NewViewName("New Name")
	err = view.Rename(newName)
	assert.Equal(t, ErrViewAlreadyDeleted, err)

	err = view.SetAsDefault()
	assert.Equal(t, ErrViewAlreadyDeleted, err)

	err = view.UnsetAsDefault()
	assert.Equal(t, ErrViewAlreadyDeleted, err)
}

func TestLoadArchitectureViewFromHistory(t *testing.T) {
	// Create a view and capture its events
	name, err := valueobjects.NewViewName("Test View")
	require.NoError(t, err)

	originalView, err := NewArchitectureView(name, "Test description", false)
	require.NoError(t, err)

	componentID := uuid.New().String()
	position := valueobjects.NewComponentPosition(150, 250)
	err = originalView.AddComponent(componentID, position)
	require.NoError(t, err)

	events := originalView.GetUncommittedChanges()

	// Reconstruct from history
	reconstructedView, err := LoadArchitectureViewFromHistory(events)

	require.NoError(t, err)
	assert.NotNil(t, reconstructedView)
	assert.Equal(t, originalView.ID(), reconstructedView.ID())
	assert.Equal(t, originalView.Name(), reconstructedView.Name())
	assert.Equal(t, originalView.Description(), reconstructedView.Description())
	assert.Equal(t, originalView.IsDefault(), reconstructedView.IsDefault())
	assert.Equal(t, originalView.IsDeleted(), reconstructedView.IsDeleted())
	assert.Len(t, reconstructedView.Components(), 1)
	assert.Contains(t, reconstructedView.Components(), componentID)
}

func TestArchitectureView_MultipleComponents(t *testing.T) {
	name, err := valueobjects.NewViewName("Test View")
	require.NoError(t, err)

	view, err := NewArchitectureView(name, "Test view", false)
	require.NoError(t, err)

	// Add multiple components
	component1ID := uuid.New().String()
	component2ID := uuid.New().String()
	component3ID := uuid.New().String()

	err = view.AddComponent(component1ID, valueobjects.NewComponentPosition(100, 100))
	require.NoError(t, err)

	err = view.AddComponent(component2ID, valueobjects.NewComponentPosition(200, 200))
	require.NoError(t, err)

	err = view.AddComponent(component3ID, valueobjects.NewComponentPosition(300, 300))
	require.NoError(t, err)

	components := view.Components()
	assert.Len(t, components, 3)
	assert.Contains(t, components, component1ID)
	assert.Contains(t, components, component2ID)
	assert.Contains(t, components, component3ID)

	// Remove one component
	err = view.RemoveComponent(component2ID)
	require.NoError(t, err)

	components = view.Components()
	assert.Len(t, components, 2)
	assert.NotContains(t, components, component2ID)
}
