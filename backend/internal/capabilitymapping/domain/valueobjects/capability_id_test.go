package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCapabilityID(t *testing.T) {
	id := NewCapabilityID()
	assert.NotEmpty(t, id.Value())
}

func TestNewCapabilityIDFromString_Valid(t *testing.T) {
	uuidStr := "550e8400-e29b-41d4-a716-446655440000"
	id, err := NewCapabilityIDFromString(uuidStr)
	assert.NoError(t, err)
	assert.Equal(t, uuidStr, id.Value())
}

func TestNewCapabilityIDFromString_Empty(t *testing.T) {
	_, err := NewCapabilityIDFromString("")
	assert.Error(t, err)
}

func TestNewCapabilityIDFromString_InvalidUUID(t *testing.T) {
	_, err := NewCapabilityIDFromString("not-a-uuid")
	assert.Error(t, err)
}

func TestCapabilityID_String(t *testing.T) {
	id := NewCapabilityID()
	assert.Equal(t, id.Value(), id.String())
}

func TestCapabilityID_Equals(t *testing.T) {
	id1, _ := NewCapabilityIDFromString("550e8400-e29b-41d4-a716-446655440000")
	id2, _ := NewCapabilityIDFromString("550e8400-e29b-41d4-a716-446655440000")
	id3, _ := NewCapabilityIDFromString("660e8400-e29b-41d4-a716-446655440000")

	assert.True(t, id1.Equals(id2))
	assert.False(t, id1.Equals(id3))
}
