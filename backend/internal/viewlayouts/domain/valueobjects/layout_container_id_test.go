package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLayoutContainerID(t *testing.T) {
	id := NewLayoutContainerID()
	assert.NotEmpty(t, id.Value())
}

func TestNewLayoutContainerIDFromString_Valid(t *testing.T) {
	uuidStr := "550e8400-e29b-41d4-a716-446655440000"
	id, err := NewLayoutContainerIDFromString(uuidStr)
	assert.NoError(t, err)
	assert.Equal(t, uuidStr, id.Value())
}

func TestNewLayoutContainerIDFromString_Empty(t *testing.T) {
	_, err := NewLayoutContainerIDFromString("")
	assert.Error(t, err)
}

func TestNewLayoutContainerIDFromString_InvalidUUID(t *testing.T) {
	_, err := NewLayoutContainerIDFromString("not-a-uuid")
	assert.Error(t, err)
}

func TestLayoutContainerID_String(t *testing.T) {
	id := NewLayoutContainerID()
	assert.Equal(t, id.Value(), id.String())
}

func TestLayoutContainerID_Equals(t *testing.T) {
	id1, _ := NewLayoutContainerIDFromString("550e8400-e29b-41d4-a716-446655440000")
	id2, _ := NewLayoutContainerIDFromString("550e8400-e29b-41d4-a716-446655440000")
	id3, _ := NewLayoutContainerIDFromString("660e8400-e29b-41d4-a716-446655440000")

	assert.True(t, id1.Equals(id2))
	assert.False(t, id1.Equals(id3))
}
