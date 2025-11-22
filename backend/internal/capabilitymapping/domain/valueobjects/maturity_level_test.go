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
		{"Genesis", "Genesis", MaturityGenesis},
		{"Custom Build", "Custom Build", MaturityCustomBuild},
		{"Product", "Product", MaturityProduct},
		{"Commodity", "Commodity", MaturityCommodity},
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
	level, err := NewMaturityLevel("  Product  ")
	assert.NoError(t, err)
	assert.Equal(t, MaturityProduct, level)
}

func TestNewMaturityLevel_Empty(t *testing.T) {
	level, err := NewMaturityLevel("")
	assert.NoError(t, err)
	assert.Equal(t, MaturityGenesis, level)
}

func TestNewMaturityLevel_InvalidValue(t *testing.T) {
	_, err := NewMaturityLevel("InvalidLevel")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidMaturityLevel, err)
}

func TestMaturityLevel_Value(t *testing.T) {
	level := MaturityProduct
	assert.Equal(t, "Product", level.Value())
}

func TestMaturityLevel_String(t *testing.T) {
	level := MaturityCommodity
	assert.Equal(t, "Commodity", level.String())
}

func TestMaturityLevel_NumericValue(t *testing.T) {
	tests := []struct {
		level    MaturityLevel
		expected int
	}{
		{MaturityGenesis, 1},
		{MaturityCustomBuild, 2},
		{MaturityProduct, 3},
		{MaturityCommodity, 4},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.level.NumericValue())
		})
	}
}

func TestMaturityLevel_Equals(t *testing.T) {
	level1 := MaturityProduct
	level2 := MaturityProduct
	level3 := MaturityGenesis

	assert.True(t, level1.Equals(level2))
	assert.False(t, level1.Equals(level3))
}
