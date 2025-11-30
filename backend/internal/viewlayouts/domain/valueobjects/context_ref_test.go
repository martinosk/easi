package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewContextRef_Valid(t *testing.T) {
	ref, err := NewContextRef("view-123")
	assert.NoError(t, err)
	assert.Equal(t, "view-123", ref.Value())
}

func TestNewContextRef_ValidUUID(t *testing.T) {
	ref, err := NewContextRef("550e8400-e29b-41d4-a716-446655440000")
	assert.NoError(t, err)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", ref.Value())
}

func TestNewContextRef_Empty(t *testing.T) {
	_, err := NewContextRef("")
	assert.Error(t, err)
	assert.Equal(t, ErrEmptyContextRef, err)
}

func TestNewContextRef_Whitespace(t *testing.T) {
	_, err := NewContextRef("   ")
	assert.Error(t, err)
	assert.Equal(t, ErrEmptyContextRef, err)
}

func TestNewContextRef_TrimSpace(t *testing.T) {
	ref, err := NewContextRef("  view-123  ")
	assert.NoError(t, err)
	assert.Equal(t, "view-123", ref.Value())
}

func TestContextRef_String(t *testing.T) {
	ref, _ := NewContextRef("domain-finance")
	assert.Equal(t, "domain-finance", ref.String())
}

func TestContextRef_Equals(t *testing.T) {
	ref1, _ := NewContextRef("view-123")
	ref2, _ := NewContextRef("view-123")
	ref3, _ := NewContextRef("view-456")

	assert.True(t, ref1.Equals(ref2))
	assert.False(t, ref1.Equals(ref3))
}
