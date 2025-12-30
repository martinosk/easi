package valueobjects

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDescription_WithValue(t *testing.T) {
	description, err := NewDescription("This is a test description")
	require.NoError(t, err)

	assert.Equal(t, "This is a test description", description.Value())
	assert.False(t, description.IsEmpty())
}

func TestNewDescription_Empty(t *testing.T) {
	description, err := NewDescription("")
	require.NoError(t, err)

	assert.Equal(t, "", description.Value())
	assert.True(t, description.IsEmpty())
}

func TestNewDescription_TrimsWhitespace(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"leading spaces", "   Description text", "Description text"},
		{"trailing spaces", "Description text   ", "Description text"},
		{"both sides", "   Description text   ", "Description text"},
		{"tabs", "\t\tDescription text\t\t", "Description text"},
		{"mixed whitespace", " \t Description text \t ", "Description text"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			description, err := NewDescription(tc.input)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, description.Value())
		})
	}
}

func TestNewDescription_WhitespaceOnly(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{"spaces", "   "},
		{"tabs", "\t\t\t"},
		{"newlines", "\n\n"},
		{"mixed", " \t \n "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			description, err := NewDescription(tc.input)
			require.NoError(t, err)

			assert.Equal(t, "", description.Value())
			assert.True(t, description.IsEmpty())
		})
	}
}

func TestDescription_IsEmpty(t *testing.T) {
	emptyDesc, err := NewDescription("")
	require.NoError(t, err)
	assert.True(t, emptyDesc.IsEmpty())

	whitespaceDesc, err := NewDescription("   ")
	require.NoError(t, err)
	assert.True(t, whitespaceDesc.IsEmpty())

	nonEmptyDesc, err := NewDescription("Some content")
	require.NoError(t, err)
	assert.False(t, nonEmptyDesc.IsEmpty())
}

func TestDescription_Equals(t *testing.T) {
	desc1, _ := NewDescription("User authentication module")
	desc2, _ := NewDescription("User authentication module")
	desc3, _ := NewDescription("Different description")
	emptyDesc1, _ := NewDescription("")
	emptyDesc2, _ := NewDescription("")

	assert.True(t, desc1.Equals(desc2))
	assert.True(t, desc2.Equals(desc1))

	assert.False(t, desc1.Equals(desc3))
	assert.False(t, desc3.Equals(desc1))

	assert.True(t, emptyDesc1.Equals(emptyDesc2))

	assert.False(t, emptyDesc1.Equals(desc1))
}

func TestDescription_Equals_WithDifferentValueObjectType(t *testing.T) {
	description, _ := NewDescription("some-uuid-value")

	uuidValue := NewUUIDValue()

	assert.False(t, description.Equals(uuidValue))
}

func TestDescription_String(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal text", "Payment processing service", "Payment processing service"},
		{"empty", "", ""},
		{"with whitespace trimmed", "  Test  ", "Test"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			description, err := NewDescription(tc.input)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, description.String())
		})
	}
}

func TestDescription_LongText(t *testing.T) {
	longText := "This is a very long description that contains multiple sentences. " +
		"It describes the component in great detail, providing context about its purpose, " +
		"functionality, and integration points. The description can span multiple lines " +
		"and include technical details about the component's implementation."

	description, err := NewDescription(longText)
	require.NoError(t, err)

	assert.Equal(t, longText, description.Value())
	assert.False(t, description.IsEmpty())
}

func TestDescription_SpecialCharacters(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{"punctuation", "Component with: colons, commas, and periods."},
		{"quotes", "Component with \"quotes\" and 'apostrophes'"},
		{"symbols", "Component with symbols @#$%^&*()"},
		{"unicode", "Component with unicode: \u00e9\u00f1\u00fc"},
		{"newlines", "Line 1\nLine 2\nLine 3"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			description, err := NewDescription(tc.input)
			require.NoError(t, err)

			assert.Contains(t, description.Value(), tc.input)
			assert.False(t, description.IsEmpty())
		})
	}
}

func TestNewDescription_TooLong(t *testing.T) {
	longText := strings.Repeat("a", MaxDescriptionLength+1)

	_, err := NewDescription(longText)

	assert.ErrorIs(t, err, ErrDescriptionTooLong)
}

func TestNewDescription_ExactMaxLength(t *testing.T) {
	exactText := strings.Repeat("a", MaxDescriptionLength)

	description, err := NewDescription(exactText)
	require.NoError(t, err)

	assert.Equal(t, exactText, description.Value())
}

func TestMustNewDescription_Valid(t *testing.T) {
	description := MustNewDescription("Valid description")

	assert.Equal(t, "Valid description", description.Value())
}

func TestMustNewDescription_PanicsOnTooLong(t *testing.T) {
	longText := strings.Repeat("a", MaxDescriptionLength+1)

	assert.Panics(t, func() {
		MustNewDescription(longText)
	})
}
