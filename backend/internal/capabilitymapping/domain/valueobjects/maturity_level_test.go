package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMaturityLevel_Valid(t *testing.T) {
	level, err := NewMaturityLevel("Product")
	assert.NoError(t, err)
	assert.Equal(t, 62, level.Value())
	assert.Equal(t, "Product", level.SectionName())
}

func TestNewMaturityLevel_Empty(t *testing.T) {
	level, err := NewMaturityLevel("")
	assert.NoError(t, err)
	assert.Equal(t, MaturityGenesis.Value(), level.Value())
}

func TestNewMaturityLevel_InvalidValue(t *testing.T) {
	_, err := NewMaturityLevel("InvalidLevel")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidMaturityLevel, err)
}

func TestNewMaturityLevelFromValue_Valid(t *testing.T) {
	level, err := NewMaturityLevelFromValue(42)
	assert.NoError(t, err)
	assert.Equal(t, 42, level.Value())
}

func TestNewMaturityLevelFromValue_ZeroIsValid(t *testing.T) {
	level, err := NewMaturityLevelFromValue(0)
	assert.NoError(t, err)
	assert.Equal(t, 0, level.Value())
}

func TestNewMaturityLevelFromValue_NinetyNineIsValid(t *testing.T) {
	level, err := NewMaturityLevelFromValue(99)
	assert.NoError(t, err)
	assert.Equal(t, 99, level.Value())
}

func TestNewMaturityLevelFromValue_NegativeIsInvalid(t *testing.T) {
	_, err := NewMaturityLevelFromValue(-1)
	assert.Error(t, err)
	assert.Equal(t, ErrMaturityValueOutOfRange, err)
}

func TestNewMaturityLevelFromValue_Over99IsInvalid(t *testing.T) {
	_, err := NewMaturityLevelFromValue(100)
	assert.Error(t, err)
	assert.Equal(t, ErrMaturityValueOutOfRange, err)
}

func TestMaturityLevel_SectionName(t *testing.T) {
	testCases := []struct {
		value        int
		expectedName string
	}{
		{0, "Genesis"},
		{12, "Genesis"},
		{24, "Genesis"},
		{25, "Custom Build"},
		{37, "Custom Build"},
		{49, "Custom Build"},
		{50, "Product"},
		{62, "Product"},
		{74, "Product"},
		{75, "Commodity"},
		{87, "Commodity"},
		{99, "Commodity"},
	}

	for _, tc := range testCases {
		level, err := NewMaturityLevelFromValue(tc.value)
		assert.NoError(t, err)
		assert.Equal(t, tc.expectedName, level.SectionName(), "value %d should have section name %s", tc.value, tc.expectedName)
	}
}

func TestMaturityLevel_SectionOrder(t *testing.T) {
	testCases := []struct {
		value         int
		expectedOrder int
	}{
		{0, 1},
		{24, 1},
		{25, 2},
		{49, 2},
		{50, 3},
		{74, 3},
		{75, 4},
		{99, 4},
	}

	for _, tc := range testCases {
		level, err := NewMaturityLevelFromValue(tc.value)
		assert.NoError(t, err)
		assert.Equal(t, tc.expectedOrder, level.SectionOrder(), "value %d should have order %d", tc.value, tc.expectedOrder)
	}
}

func TestMaturityLevel_SectionRange(t *testing.T) {
	testCases := []struct {
		value       int
		expectedMin int
		expectedMax int
	}{
		{12, 0, 24},
		{37, 25, 49},
		{62, 50, 74},
		{87, 75, 99},
	}

	for _, tc := range testCases {
		level, err := NewMaturityLevelFromValue(tc.value)
		assert.NoError(t, err)
		min, max := level.SectionRange()
		assert.Equal(t, tc.expectedMin, min)
		assert.Equal(t, tc.expectedMax, max)
	}
}

func TestMaturityLevel_NumericValue(t *testing.T) {
	assert.Equal(t, 1, MaturityGenesis.NumericValue())
	assert.Equal(t, 2, MaturityCustomBuild.NumericValue())
	assert.Equal(t, 3, MaturityProduct.NumericValue())
	assert.Equal(t, 4, MaturityCommodity.NumericValue())
}

func TestMaturityLevel_Equals(t *testing.T) {
	level1 := MaturityProduct
	level2 := MaturityProduct
	level3 := MaturityGenesis

	assert.True(t, level1.Equals(level2))
	assert.False(t, level1.Equals(level3))
}

func TestMaturityLevel_LegacyStringConversion(t *testing.T) {
	testCases := []struct {
		input         string
		expectedValue int
	}{
		{"Genesis", 12},
		{"Custom Build", 37},
		{"Product", 62},
		{"Commodity", 87},
	}

	for _, tc := range testCases {
		level, err := NewMaturityLevel(tc.input)
		assert.NoError(t, err)
		assert.Equal(t, tc.expectedValue, level.Value(), "legacy string %s should convert to %d", tc.input, tc.expectedValue)
	}
}
