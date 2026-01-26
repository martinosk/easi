package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFitType_ValidValues(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"technical", "TECHNICAL", "TECHNICAL"},
		{"functional", "FUNCTIONAL", "FUNCTIONAL"},
		{"empty string", "", ""},
		{"lowercase technical", "technical", "TECHNICAL"},
		{"lowercase functional", "functional", "FUNCTIONAL"},
		{"mixed case technical", "Technical", "TECHNICAL"},
		{"with leading whitespace", "  TECHNICAL", "TECHNICAL"},
		{"with trailing whitespace", "FUNCTIONAL  ", "FUNCTIONAL"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ft, err := NewFitType(tc.input)

			require.NoError(t, err)
			assert.Equal(t, tc.expected, ft.Value())
		})
	}
}

func TestNewFitType_InvalidValues(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{"invalid value", "INVALID"},
		{"typo", "TECHNCAL"},
		{"partial", "TECH"},
		{"random string", "foo"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewFitType(tc.input)

			assert.Error(t, err)
			assert.Equal(t, ErrInvalidFitType, err)
		})
	}
}

func TestFitType_IsEmpty(t *testing.T) {
	emptyFit, _ := NewFitType("")
	technicalFit, _ := NewFitType("TECHNICAL")
	functionalFit, _ := NewFitType("FUNCTIONAL")

	assert.True(t, emptyFit.IsEmpty())
	assert.False(t, technicalFit.IsEmpty())
	assert.False(t, functionalFit.IsEmpty())
}

func TestFitType_IsTechnical(t *testing.T) {
	emptyFit, _ := NewFitType("")
	technicalFit, _ := NewFitType("TECHNICAL")
	functionalFit, _ := NewFitType("FUNCTIONAL")

	assert.False(t, emptyFit.IsTechnical())
	assert.True(t, technicalFit.IsTechnical())
	assert.False(t, functionalFit.IsTechnical())
}

func TestFitType_IsFunctional(t *testing.T) {
	emptyFit, _ := NewFitType("")
	technicalFit, _ := NewFitType("TECHNICAL")
	functionalFit, _ := NewFitType("FUNCTIONAL")

	assert.False(t, emptyFit.IsFunctional())
	assert.False(t, technicalFit.IsFunctional())
	assert.True(t, functionalFit.IsFunctional())
}

func TestFitType_Equals(t *testing.T) {
	tech1, _ := NewFitType("TECHNICAL")
	tech2, _ := NewFitType("technical")
	func1, _ := NewFitType("FUNCTIONAL")
	empty, _ := NewFitType("")

	assert.True(t, tech1.Equals(tech2))
	assert.False(t, tech1.Equals(func1))
	assert.False(t, tech1.Equals(empty))
	assert.True(t, empty.Equals(empty))
}

func TestFitType_String(t *testing.T) {
	tech, _ := NewFitType("TECHNICAL")
	func1, _ := NewFitType("FUNCTIONAL")
	empty, _ := NewFitType("")

	assert.Equal(t, "TECHNICAL", tech.String())
	assert.Equal(t, "FUNCTIONAL", func1.String())
	assert.Equal(t, "", empty.String())
}
