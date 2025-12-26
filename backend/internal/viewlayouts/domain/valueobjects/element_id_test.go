package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestElementID_Equals(t *testing.T) {
	id1, _ := NewElementID("comp-1")
	id2, _ := NewElementID("comp-1")
	id3, _ := NewElementID("comp-2")

	assert.True(t, id1.Equals(id2))
	assert.False(t, id1.Equals(id3))
}
