package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCapabilityName_Empty(t *testing.T) {
	_, err := NewCapabilityName("")
	assert.Error(t, err)
	assert.Equal(t, ErrCapabilityNameEmpty, err)
}

func TestNewCapabilityName_OnlyWhitespace(t *testing.T) {
	_, err := NewCapabilityName("   ")
	assert.Error(t, err)
	assert.Equal(t, ErrCapabilityNameEmpty, err)
}
