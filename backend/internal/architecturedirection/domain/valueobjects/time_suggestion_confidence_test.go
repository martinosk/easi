package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTimeSuggestionConfidence_ValidValues(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"low", "LOW", "LOW"},
		{"medium", "MEDIUM", "MEDIUM"},
		{"high", "HIGH", "HIGH"},
		{"lowercase", "low", "LOW"},
		{"mixed case", "Medium", "MEDIUM"},
		{"with whitespace", "  HIGH  ", "HIGH"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conf, err := NewTimeSuggestionConfidence(tc.input)

			require.NoError(t, err)
			assert.Equal(t, tc.expected, conf.Value())
		})
	}
}

func TestNewTimeSuggestionConfidence_InvalidValues(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"invalid", "INVALID"},
		{"typo", "HIHG"},
		{"partial", "MED"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewTimeSuggestionConfidence(tc.input)

			assert.Error(t, err)
			assert.Equal(t, ErrInvalidTimeSuggestionConfidence, err)
		})
	}
}

func TestTimeSuggestionConfidence_Predicates(t *testing.T) {
	low, _ := NewTimeSuggestionConfidence("LOW")
	medium, _ := NewTimeSuggestionConfidence("MEDIUM")
	high, _ := NewTimeSuggestionConfidence("HIGH")

	assert.True(t, low.IsLow())
	assert.False(t, low.IsMedium())
	assert.False(t, low.IsHigh())

	assert.False(t, medium.IsLow())
	assert.True(t, medium.IsMedium())
	assert.False(t, medium.IsHigh())

	assert.False(t, high.IsLow())
	assert.False(t, high.IsMedium())
	assert.True(t, high.IsHigh())
}

func TestTimeSuggestionConfidence_Equals(t *testing.T) {
	conf1, _ := NewTimeSuggestionConfidence("HIGH")
	conf2, _ := NewTimeSuggestionConfidence("high")
	conf3, _ := NewTimeSuggestionConfidence("LOW")

	assert.True(t, conf1.Equals(conf2))
	assert.False(t, conf1.Equals(conf3))
}

func TestTimeSuggestionConfidence_String(t *testing.T) {
	high, _ := NewTimeSuggestionConfidence("HIGH")
	assert.Equal(t, "HIGH", high.String())
}
