package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewElementPosition_Valid(t *testing.T) {
	elementID, _ := NewElementID("component-123")
	pos, err := NewElementPosition(elementID, 100.5, 200.25)
	assert.NoError(t, err)
	assert.Equal(t, "component-123", pos.ElementID().Value())
	assert.Equal(t, 100.5, pos.X())
	assert.Equal(t, 200.25, pos.Y())
	assert.Nil(t, pos.Width())
	assert.Nil(t, pos.Height())
	assert.Nil(t, pos.CustomColor())
	assert.Nil(t, pos.SortOrder())
}

func TestNewElementPosition_WithOptionalFields(t *testing.T) {
	elementID, _ := NewElementID("cap-456")
	width := 150.0
	height := 100.0
	color, _ := NewHexColor("#3b82f6")
	sortOrder := 5

	pos, err := NewElementPositionWithOptions(elementID, 50, 75, &width, &height, &color, &sortOrder)
	assert.NoError(t, err)
	assert.Equal(t, 50.0, pos.X())
	assert.Equal(t, 75.0, pos.Y())
	assert.Equal(t, 150.0, *pos.Width())
	assert.Equal(t, 100.0, *pos.Height())
	assert.Equal(t, "#3b82f6", pos.CustomColor().Value())
	assert.Equal(t, 5, *pos.SortOrder())
}

func TestNewElementPosition_NegativeCoordinates(t *testing.T) {
	elementID, _ := NewElementID("component-123")
	pos, err := NewElementPosition(elementID, -100, -50)
	assert.NoError(t, err)
	assert.Equal(t, -100.0, pos.X())
	assert.Equal(t, -50.0, pos.Y())
}

func TestElementPosition_Equals(t *testing.T) {
	elementID1, _ := NewElementID("comp-1")
	elementID2, _ := NewElementID("comp-1")
	elementID3, _ := NewElementID("comp-2")

	pos1, _ := NewElementPosition(elementID1, 100, 200)
	pos2, _ := NewElementPosition(elementID2, 100, 200)
	pos3, _ := NewElementPosition(elementID3, 100, 200)
	pos4, _ := NewElementPosition(elementID1, 150, 200)

	assert.True(t, pos1.Equals(pos2))
	assert.False(t, pos1.Equals(pos3))
	assert.False(t, pos1.Equals(pos4))
}

func TestElementPosition_WithUpdatedPosition(t *testing.T) {
	elementID, _ := NewElementID("component-123")
	pos, _ := NewElementPosition(elementID, 100, 200)

	newPos := pos.WithUpdatedPosition(300, 400)
	assert.Equal(t, 300.0, newPos.X())
	assert.Equal(t, 400.0, newPos.Y())
	assert.Equal(t, pos.ElementID(), newPos.ElementID())
}
