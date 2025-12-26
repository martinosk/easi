package valueobjects

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSectionName_ValidNames(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple name", "Genesis", "Genesis"},
		{"with spaces", "Custom Built", "Custom Built"},
		{"trims whitespace", "  Product  ", "Product"},
		{"single char", "A", "A"},
		{"max length 50", strings.Repeat("a", 50), strings.Repeat("a", 50)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sn, err := NewSectionName(tc.input)

			require.NoError(t, err)
			assert.Equal(t, tc.expected, sn.Value())
		})
	}
}

func TestNewSectionName_InvalidNames(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectedErr error
	}{
		{"empty string", "", ErrSectionNameEmpty},
		{"whitespace only", "   ", ErrSectionNameEmpty},
		{"too long 51 chars", strings.Repeat("a", 51), ErrSectionNameTooLong},
		{"way too long", strings.Repeat("x", 100), ErrSectionNameTooLong},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewSectionName(tc.input)

			assert.Error(t, err)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestSectionName_Equals(t *testing.T) {
	sn1, _ := NewSectionName("Genesis")
	sn2, _ := NewSectionName("Genesis")
	sn3, _ := NewSectionName("Product")

	assert.True(t, sn1.Equals(sn2))
	assert.False(t, sn1.Equals(sn3))
}

func TestSectionName_String(t *testing.T) {
	sn, _ := NewSectionName("Custom Built")

	assert.Equal(t, "Custom Built", sn.String())
}
