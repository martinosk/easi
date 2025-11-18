package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCapabilityName_Valid(t *testing.T) {
	name, err := NewCapabilityName("Customer Management")
	assert.NoError(t, err)
	assert.Equal(t, "Customer Management", name.Value())
}

func TestNewCapabilityName_TrimSpace(t *testing.T) {
	name, err := NewCapabilityName("  Customer Management  ")
	assert.NoError(t, err)
	assert.Equal(t, "Customer Management", name.Value())
}

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

func TestCapabilityName_String(t *testing.T) {
	name, _ := NewCapabilityName("Digital Transformation")
	assert.Equal(t, "Digital Transformation", name.String())
}
