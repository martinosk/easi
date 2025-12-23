package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDescription_WithValue(t *testing.T) {
	description := NewDescription("This is a test description")

	assert.Equal(t, "This is a test description", description.Value())
	assert.False(t, description.IsEmpty())
}

func TestNewDescription_Empty(t *testing.T) {
	description := NewDescription("")

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
			description := NewDescription(tc.input)

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
			description := NewDescription(tc.input)

			assert.Equal(t, "", description.Value())
			assert.True(t, description.IsEmpty())
		})
	}
}

func TestDescription_IsEmpty(t *testing.T) {
	emptyDesc := NewDescription("")
	assert.True(t, emptyDesc.IsEmpty())

	whitespaceDesc := NewDescription("   ")
	assert.True(t, whitespaceDesc.IsEmpty())

	nonEmptyDesc := NewDescription("Some content")
	assert.False(t, nonEmptyDesc.IsEmpty())
}

func TestDescription_Equals(t *testing.T) {
	desc1 := NewDescription("User authentication module")
	desc2 := NewDescription("User authentication module")
	desc3 := NewDescription("Different description")
	emptyDesc1 := NewDescription("")
	emptyDesc2 := NewDescription("")

	assert.True(t, desc1.Equals(desc2))
	assert.True(t, desc2.Equals(desc1))

	assert.False(t, desc1.Equals(desc3))
	assert.False(t, desc3.Equals(desc1))

	assert.True(t, emptyDesc1.Equals(emptyDesc2))

	assert.False(t, emptyDesc1.Equals(desc1))
}

func TestDescription_Equals_WithDifferentValueObjectType(t *testing.T) {
	description := NewDescription("some-uuid-value")

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
			description := NewDescription(tc.input)

			assert.Equal(t, tc.expected, description.String())
		})
	}
}

func TestDescription_LongText(t *testing.T) {
	longText := "This is a very long description that contains multiple sentences. " +
		"It describes the component in great detail, providing context about its purpose, " +
		"functionality, and integration points. The description can span multiple lines " +
		"and include technical details about the component's implementation."

	description := NewDescription(longText)

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
			description := NewDescription(tc.input)

			assert.Contains(t, description.Value(), tc.input)
			assert.False(t, description.IsEmpty())
		})
	}
}
