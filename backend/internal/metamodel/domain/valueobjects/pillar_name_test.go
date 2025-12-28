package valueobjects

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPillarName_ValidNames(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple name", "Always On", "Always On"},
		{"with spaces", "Digital First", "Digital First"},
		{"trims whitespace", "  Innovation  ", "Innovation"},
		{"single char", "A", "A"},
		{"max length 100", strings.Repeat("a", 100), strings.Repeat("a", 100)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pn, err := NewPillarName(tc.input)

			require.NoError(t, err)
			assert.Equal(t, tc.expected, pn.Value())
		})
	}
}

func TestNewPillarName_InvalidNames(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectedErr error
	}{
		{"empty string", "", ErrPillarNameEmpty},
		{"whitespace only", "   ", ErrPillarNameEmpty},
		{"too long 101 chars", strings.Repeat("a", 101), ErrPillarNameTooLong},
		{"way too long", strings.Repeat("x", 200), ErrPillarNameTooLong},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewPillarName(tc.input)

			assert.Error(t, err)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestPillarName_EqualsIgnoreCase(t *testing.T) {
	pn1, _ := NewPillarName("Always On")
	pn2, _ := NewPillarName("always on")
	pn3, _ := NewPillarName("ALWAYS ON")
	pn4, _ := NewPillarName("Different")

	assert.True(t, pn1.EqualsIgnoreCase(pn2))
	assert.True(t, pn1.EqualsIgnoreCase(pn3))
	assert.True(t, pn2.EqualsIgnoreCase(pn3))
	assert.False(t, pn1.EqualsIgnoreCase(pn4))
}
