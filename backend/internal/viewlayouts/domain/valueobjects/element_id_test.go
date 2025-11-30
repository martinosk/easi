package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewElementID_Valid(t *testing.T) {
	id, err := NewElementID("component-123")
	assert.NoError(t, err)
	assert.Equal(t, "component-123", id.Value())
}

func TestNewElementID_ValidUUID(t *testing.T) {
	id, err := NewElementID("550e8400-e29b-41d4-a716-446655440000")
	assert.NoError(t, err)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", id.Value())
}

func TestNewElementID_Empty(t *testing.T) {
	_, err := NewElementID("")
	assert.Error(t, err)
	assert.Equal(t, ErrEmptyElementID, err)
}

func TestNewElementID_Whitespace(t *testing.T) {
	_, err := NewElementID("   ")
	assert.Error(t, err)
	assert.Equal(t, ErrEmptyElementID, err)
}

func TestNewElementID_TrimSpace(t *testing.T) {
	id, err := NewElementID("  component-123  ")
	assert.NoError(t, err)
	assert.Equal(t, "component-123", id.Value())
}

func TestElementID_String(t *testing.T) {
	id, _ := NewElementID("cap-456")
	assert.Equal(t, "cap-456", id.String())
}

func TestElementID_Equals(t *testing.T) {
	id1, _ := NewElementID("comp-1")
	id2, _ := NewElementID("comp-1")
	id3, _ := NewElementID("comp-2")

	assert.True(t, id1.Equals(id2))
	assert.False(t, id1.Equals(id3))
}
