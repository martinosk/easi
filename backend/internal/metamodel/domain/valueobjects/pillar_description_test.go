package valueobjects

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPillarDescription_ValidDescriptions(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple description", "Core capabilities", "Core capabilities"},
		{"empty is valid", "", ""},
		{"trims whitespace", "  Description  ", "Description"},
		{"max length 500", strings.Repeat("a", 500), strings.Repeat("a", 500)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pd, err := NewPillarDescription(tc.input)

			require.NoError(t, err)
			assert.Equal(t, tc.expected, pd.Value())
		})
	}
}

func TestNewPillarDescription_InvalidDescriptions(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectedErr error
	}{
		{"too long 501 chars", strings.Repeat("a", 501), ErrPillarDescriptionTooLong},
		{"way too long", strings.Repeat("x", 1000), ErrPillarDescriptionTooLong},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewPillarDescription(tc.input)

			assert.Error(t, err)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestPillarDescription_IsEmpty(t *testing.T) {
	empty, _ := NewPillarDescription("")
	whitespaceOnly, _ := NewPillarDescription("   ")
	nonEmpty, _ := NewPillarDescription("Something")

	assert.True(t, empty.IsEmpty())
	assert.True(t, whitespaceOnly.IsEmpty())
	assert.False(t, nonEmpty.IsEmpty())
}
