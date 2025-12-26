package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCapabilityIDFromString_Empty(t *testing.T) {
	_, err := NewCapabilityIDFromString("")
	assert.Error(t, err)
}

func TestCapabilityID_Equals(t *testing.T) {
	id1, _ := NewCapabilityIDFromString("550e8400-e29b-41d4-a716-446655440000")
	id2, _ := NewCapabilityIDFromString("550e8400-e29b-41d4-a716-446655440000")
	id3, _ := NewCapabilityIDFromString("660e8400-e29b-41d4-a716-446655440000")

	assert.True(t, id1.Equals(id2))
	assert.False(t, id1.Equals(id3))
}
