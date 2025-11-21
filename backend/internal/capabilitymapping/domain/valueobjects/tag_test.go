package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTag_ValidValue(t *testing.T) {
	tag, err := NewTag("Legacy")
	assert.NoError(t, err)
	assert.Equal(t, "Legacy", tag.Value())
}

func TestNewTag_TrimSpace(t *testing.T) {
	tag, err := NewTag("  API-first  ")
	assert.NoError(t, err)
	assert.Equal(t, "API-first", tag.Value())
}

func TestNewTag_Empty(t *testing.T) {
	_, err := NewTag("")
	assert.Error(t, err)
	assert.Equal(t, ErrTagEmpty, err)
}

func TestNewTag_OnlyWhitespace(t *testing.T) {
	_, err := NewTag("   ")
	assert.Error(t, err)
	assert.Equal(t, ErrTagEmpty, err)
}

func TestTag_Value(t *testing.T) {
	tag, _ := NewTag("Cloud-native")
	assert.Equal(t, "Cloud-native", tag.Value())
}

func TestTag_String(t *testing.T) {
	tag, _ := NewTag("Microservices")
	assert.Equal(t, "Microservices", tag.String())
}

func TestTag_Equals(t *testing.T) {
	tag1, _ := NewTag("Legacy")
	tag2, _ := NewTag("Legacy")
	tag3, _ := NewTag("Modern")

	assert.True(t, tag1.Equals(tag2))
	assert.False(t, tag1.Equals(tag3))
}
