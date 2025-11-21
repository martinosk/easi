package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMaturityLevel_ValidValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected MaturityLevel
	}{
		{"Initial", "Initial", MaturityInitial},
		{"Developing", "Developing", MaturityDeveloping},
		{"Defined", "Defined", MaturityDefined},
		{"Managed", "Managed", MaturityManaged},
		{"Optimizing", "Optimizing", MaturityOptimizing},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, err := NewMaturityLevel(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, level)
		})
	}
}

func TestNewMaturityLevel_TrimSpace(t *testing.T) {
	level, err := NewMaturityLevel("  Managed  ")
	assert.NoError(t, err)
	assert.Equal(t, MaturityManaged, level)
}

func TestNewMaturityLevel_Empty(t *testing.T) {
	level, err := NewMaturityLevel("")
	assert.NoError(t, err)
	assert.Equal(t, MaturityInitial, level)
}

func TestNewMaturityLevel_InvalidValue(t *testing.T) {
	_, err := NewMaturityLevel("InvalidLevel")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidMaturityLevel, err)
}

func TestMaturityLevel_Value(t *testing.T) {
	level := MaturityDefined
	assert.Equal(t, "Defined", level.Value())
}

func TestMaturityLevel_String(t *testing.T) {
	level := MaturityOptimizing
	assert.Equal(t, "Optimizing", level.String())
}

func TestMaturityLevel_NumericValue(t *testing.T) {
	tests := []struct {
		level    MaturityLevel
		expected int
	}{
		{MaturityInitial, 1},
		{MaturityDeveloping, 2},
		{MaturityDefined, 3},
		{MaturityManaged, 4},
		{MaturityOptimizing, 5},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.level.NumericValue())
		})
	}
}

func TestMaturityLevel_Equals(t *testing.T) {
	level1 := MaturityManaged
	level2 := MaturityManaged
	level3 := MaturityDefined

	assert.True(t, level1.Equals(level2))
	assert.False(t, level1.Equals(level3))
}
