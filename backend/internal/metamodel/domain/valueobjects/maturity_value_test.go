package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMaturityValue_ValidValues(t *testing.T) {
	testCases := []struct {
		name  string
		value int
	}{
		{"minimum value", 0},
		{"maximum value", 99},
		{"mid value", 50},
		{"low value", 1},
		{"high value", 98},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mv, err := NewMaturityValue(tc.value)

			require.NoError(t, err)
			assert.Equal(t, tc.value, mv.Value())
		})
	}
}

func TestNewMaturityValue_InvalidValues(t *testing.T) {
	testCases := []struct {
		name  string
		value int
	}{
		{"negative value", -1},
		{"too high", 100},
		{"very negative", -100},
		{"very high", 1000},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewMaturityValue(tc.value)

			assert.Error(t, err)
			assert.Equal(t, ErrMaturityValueOutOfRange, err)
		})
	}
}

func TestMaturityValue_Equals(t *testing.T) {
	mv1, _ := NewMaturityValue(50)
	mv2, _ := NewMaturityValue(50)
	mv3, _ := NewMaturityValue(75)

	assert.True(t, mv1.Equals(mv2))
	assert.False(t, mv1.Equals(mv3))
}

func TestMaturityValue_String(t *testing.T) {
	mv, _ := NewMaturityValue(42)

	assert.Equal(t, "42", mv.String())
}

func TestMaturityValue_LessThan(t *testing.T) {
	mv1, _ := NewMaturityValue(25)
	mv2, _ := NewMaturityValue(50)

	assert.True(t, mv1.LessThan(mv2))
	assert.False(t, mv2.LessThan(mv1))
	assert.False(t, mv1.LessThan(mv1))
}

func TestMaturityValue_LessThanOrEqual(t *testing.T) {
	mv1, _ := NewMaturityValue(25)
	mv2, _ := NewMaturityValue(50)

	assert.True(t, mv1.LessThanOrEqual(mv2))
	assert.True(t, mv1.LessThanOrEqual(mv1))
	assert.False(t, mv2.LessThanOrEqual(mv1))
}
