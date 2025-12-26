package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCapabilityLevel_Valid(t *testing.T) {
	level, err := NewCapabilityLevel("L2")
	assert.NoError(t, err)
	assert.Equal(t, LevelL2, level)
	assert.Equal(t, 2, level.NumericValue())
}

func TestNewCapabilityLevel_CaseInsensitive(t *testing.T) {
	level, err := NewCapabilityLevel("l2")
	assert.NoError(t, err)
	assert.Equal(t, LevelL2, level)
}

func TestNewCapabilityLevel_Invalid(t *testing.T) {
	_, err := NewCapabilityLevel("L5")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCapabilityLevel, err)
}

func TestNewCapabilityLevel_Empty(t *testing.T) {
	_, err := NewCapabilityLevel("")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCapabilityLevel, err)
}

func TestCapabilityLevel_IsValid(t *testing.T) {
	level, _ := NewCapabilityLevel("L1")
	assert.True(t, level.IsValid())

	invalidLevel := CapabilityLevel("L5")
	assert.False(t, invalidLevel.IsValid())
}
