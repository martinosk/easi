package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCapabilityLevel_L1(t *testing.T) {
	level, err := NewCapabilityLevel("L1")
	assert.NoError(t, err)
	assert.Equal(t, LevelL1, level)
	assert.Equal(t, 1, level.NumericValue())
}

func TestNewCapabilityLevel_L2(t *testing.T) {
	level, err := NewCapabilityLevel("L2")
	assert.NoError(t, err)
	assert.Equal(t, LevelL2, level)
	assert.Equal(t, 2, level.NumericValue())
}

func TestNewCapabilityLevel_L3(t *testing.T) {
	level, err := NewCapabilityLevel("L3")
	assert.NoError(t, err)
	assert.Equal(t, LevelL3, level)
	assert.Equal(t, 3, level.NumericValue())
}

func TestNewCapabilityLevel_L4(t *testing.T) {
	level, err := NewCapabilityLevel("L4")
	assert.NoError(t, err)
	assert.Equal(t, LevelL4, level)
	assert.Equal(t, 4, level.NumericValue())
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

func TestNewCapabilityLevel_Invalid_Text(t *testing.T) {
	_, err := NewCapabilityLevel("Level1")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCapabilityLevel, err)
}

func TestCapabilityLevel_IsValid(t *testing.T) {
	level, _ := NewCapabilityLevel("L1")
	assert.True(t, level.IsValid())

	invalidLevel := CapabilityLevel("L5")
	assert.False(t, invalidLevel.IsValid())
}

func TestCapabilityLevel_String(t *testing.T) {
	level, _ := NewCapabilityLevel("L3")
	assert.Equal(t, "L3", level.String())
}

func TestCapabilityLevel_Value(t *testing.T) {
	level, _ := NewCapabilityLevel("L4")
	assert.Equal(t, "L4", level.Value())
}
