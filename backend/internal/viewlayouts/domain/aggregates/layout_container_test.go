package aggregates

import (
	"testing"
	"time"

	"easi/backend/internal/viewlayouts/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLayoutContainer(t *testing.T) {
	contextType, _ := valueobjects.NewLayoutContextType("business-domain-grid")
	contextRef, _ := valueobjects.NewContextRef("domain-finance")
	prefs := valueobjects.NewLayoutPreferences(nil)

	container, err := NewLayoutContainer(contextType, contextRef, prefs)
	require.NoError(t, err)

	assert.NotEmpty(t, container.ID().Value())
	assert.Equal(t, contextType, container.ContextType())
	assert.Equal(t, contextRef, container.ContextRef())
	assert.Equal(t, 1, container.Version())
	assert.NotZero(t, container.CreatedAt())
	assert.NotZero(t, container.UpdatedAt())
	assert.Empty(t, container.Elements())
}

func TestLayoutContainer_WithID(t *testing.T) {
	id, _ := valueobjects.NewLayoutContainerIDFromString("550e8400-e29b-41d4-a716-446655440000")
	contextType, _ := valueobjects.NewLayoutContextType("architecture-canvas")
	contextRef, _ := valueobjects.NewContextRef("view-123")
	prefs := valueobjects.NewLayoutPreferences(map[string]interface{}{"colorScheme": "pastel"})
	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now()

	container := NewLayoutContainerWithState(id, contextType, contextRef, prefs, 5, createdAt, updatedAt)

	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", container.ID().Value())
	assert.Equal(t, contextType, container.ContextType())
	assert.Equal(t, contextRef, container.ContextRef())
	assert.Equal(t, 5, container.Version())
	assert.Equal(t, "pastel", container.Preferences().Get("colorScheme"))
}

func TestLayoutContainer_UpsertElement(t *testing.T) {
	contextType, _ := valueobjects.NewLayoutContextType("business-domain-grid")
	contextRef, _ := valueobjects.NewContextRef("domain-finance")
	container, _ := NewLayoutContainer(contextType, contextRef, valueobjects.NewLayoutPreferences(nil))

	elementID, _ := valueobjects.NewElementID("cap-123")
	pos, _ := valueobjects.NewElementPosition(elementID, 100, 200)

	err := container.UpsertElement(pos)
	require.NoError(t, err)

	elements := container.Elements()
	assert.Len(t, elements, 1)
	assert.Equal(t, 100.0, elements[0].X())
	assert.Equal(t, 200.0, elements[0].Y())
}

func TestLayoutContainer_UpsertElement_Update(t *testing.T) {
	contextType, _ := valueobjects.NewLayoutContextType("business-domain-grid")
	contextRef, _ := valueobjects.NewContextRef("domain-finance")
	container, _ := NewLayoutContainer(contextType, contextRef, valueobjects.NewLayoutPreferences(nil))

	elementID, _ := valueobjects.NewElementID("cap-123")
	pos1, _ := valueobjects.NewElementPosition(elementID, 100, 200)
	pos2, _ := valueobjects.NewElementPosition(elementID, 300, 400)

	container.UpsertElement(pos1)
	container.UpsertElement(pos2)

	elements := container.Elements()
	assert.Len(t, elements, 1)
	assert.Equal(t, 300.0, elements[0].X())
	assert.Equal(t, 400.0, elements[0].Y())
}

func TestLayoutContainer_RemoveElement(t *testing.T) {
	contextType, _ := valueobjects.NewLayoutContextType("business-domain-grid")
	contextRef, _ := valueobjects.NewContextRef("domain-finance")
	container, _ := NewLayoutContainer(contextType, contextRef, valueobjects.NewLayoutPreferences(nil))

	elementID, _ := valueobjects.NewElementID("cap-123")
	pos, _ := valueobjects.NewElementPosition(elementID, 100, 200)
	container.UpsertElement(pos)

	err := container.RemoveElement(elementID)
	require.NoError(t, err)

	assert.Empty(t, container.Elements())
}

func TestLayoutContainer_RemoveElement_NotFound(t *testing.T) {
	contextType, _ := valueobjects.NewLayoutContextType("business-domain-grid")
	contextRef, _ := valueobjects.NewContextRef("domain-finance")
	container, _ := NewLayoutContainer(contextType, contextRef, valueobjects.NewLayoutPreferences(nil))

	elementID, _ := valueobjects.NewElementID("cap-123")
	err := container.RemoveElement(elementID)
	assert.NoError(t, err)
}

func TestLayoutContainer_UpdatePreferences(t *testing.T) {
	contextType, _ := valueobjects.NewLayoutContextType("business-domain-grid")
	contextRef, _ := valueobjects.NewContextRef("domain-finance")
	prefs := valueobjects.NewLayoutPreferences(map[string]interface{}{"colorScheme": "default"})
	container, _ := NewLayoutContainer(contextType, contextRef, prefs)

	newPrefs := valueobjects.NewLayoutPreferences(map[string]interface{}{
		"colorScheme":     "pastel",
		"layoutDirection": "LR",
	})

	err := container.UpdatePreferences(newPrefs)
	require.NoError(t, err)

	assert.Equal(t, "pastel", container.Preferences().Get("colorScheme"))
	assert.Equal(t, "LR", container.Preferences().Get("layoutDirection"))
}

func TestLayoutContainer_GetElement(t *testing.T) {
	contextType, _ := valueobjects.NewLayoutContextType("business-domain-grid")
	contextRef, _ := valueobjects.NewContextRef("domain-finance")
	container, _ := NewLayoutContainer(contextType, contextRef, valueobjects.NewLayoutPreferences(nil))

	elementID, _ := valueobjects.NewElementID("cap-123")
	pos, _ := valueobjects.NewElementPosition(elementID, 100, 200)
	container.UpsertElement(pos)

	found := container.GetElement(elementID)
	require.NotNil(t, found)
	assert.Equal(t, 100.0, found.X())
	assert.Equal(t, 200.0, found.Y())
}

func TestLayoutContainer_GetElement_NotFound(t *testing.T) {
	contextType, _ := valueobjects.NewLayoutContextType("business-domain-grid")
	contextRef, _ := valueobjects.NewContextRef("domain-finance")
	container, _ := NewLayoutContainer(contextType, contextRef, valueobjects.NewLayoutPreferences(nil))

	elementID, _ := valueobjects.NewElementID("cap-123")
	found := container.GetElement(elementID)
	assert.Nil(t, found)
}

func TestLayoutContainer_IncrementVersion(t *testing.T) {
	contextType, _ := valueobjects.NewLayoutContextType("business-domain-grid")
	contextRef, _ := valueobjects.NewContextRef("domain-finance")
	container, _ := NewLayoutContainer(contextType, contextRef, valueobjects.NewLayoutPreferences(nil))

	assert.Equal(t, 1, container.Version())
	container.IncrementVersion()
	assert.Equal(t, 2, container.Version())
}
