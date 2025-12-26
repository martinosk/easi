package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestTag_Equals(t *testing.T) {
	tag1, _ := NewTag("Legacy")
	tag2, _ := NewTag("Legacy")
	tag3, _ := NewTag("Modern")

	assert.True(t, tag1.Equals(tag2))
	assert.False(t, tag1.Equals(tag3))
}
