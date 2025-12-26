package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAssignmentIDFromString_Empty(t *testing.T) {
	_, err := NewAssignmentIDFromString("")
	assert.Error(t, err)
}

func TestNewAssignmentIDFromString_MissingPrefix(t *testing.T) {
	_, err := NewAssignmentIDFromString("550e8400-e29b-41d4-a716-446655440000")
	assert.Error(t, err)
	assert.Equal(t, ErrAssignmentIDMissingPrefix, err)
}

func TestAssignmentID_Equals(t *testing.T) {
	id1, _ := NewAssignmentIDFromString("assign-550e8400-e29b-41d4-a716-446655440000")
	id2, _ := NewAssignmentIDFromString("assign-550e8400-e29b-41d4-a716-446655440000")
	id3, _ := NewAssignmentIDFromString("assign-660e8400-e29b-41d4-a716-446655440000")

	assert.True(t, id1.Equals(id2))
	assert.False(t, id1.Equals(id3))
}
