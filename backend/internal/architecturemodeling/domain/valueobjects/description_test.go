package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDescription_WithValue(t *testing.T) {
	// Act
	description := NewDescription("This is a test description")

	// Assert
	assert.Equal(t, "This is a test description", description.Value())
	assert.False(t, description.IsEmpty())
}

func TestNewDescription_Empty(t *testing.T) {
	// Act
	description := NewDescription("")

	// Assert: Empty description is allowed
	assert.Equal(t, "", description.Value())
	assert.True(t, description.IsEmpty())
}

func TestNewDescription_TrimsWhitespace(t *testing.T) {
	// Arrange: Description with leading and trailing whitespace
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
			// Act
			description := NewDescription(tc.input)

			// Assert
			assert.Equal(t, tc.expected, description.Value())
		})
	}
}

func TestNewDescription_WhitespaceOnly(t *testing.T) {
	// Arrange: Test various whitespace-only inputs
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
			// Act
			description := NewDescription(tc.input)

			// Assert: Whitespace-only becomes empty
			assert.Equal(t, "", description.Value())
			assert.True(t, description.IsEmpty())
		})
	}
}

func TestDescription_IsEmpty(t *testing.T) {
	// Arrange & Act & Assert
	emptyDesc := NewDescription("")
	assert.True(t, emptyDesc.IsEmpty())

	whitespaceDesc := NewDescription("   ")
	assert.True(t, whitespaceDesc.IsEmpty())

	nonEmptyDesc := NewDescription("Some content")
	assert.False(t, nonEmptyDesc.IsEmpty())
}

func TestDescription_Equals(t *testing.T) {
	// Arrange
	desc1 := NewDescription("User authentication module")
	desc2 := NewDescription("User authentication module")
	desc3 := NewDescription("Different description")
	emptyDesc1 := NewDescription("")
	emptyDesc2 := NewDescription("")

	// Act & Assert: Same descriptions should be equal
	assert.True(t, desc1.Equals(desc2))
	assert.True(t, desc2.Equals(desc1))

	// Different descriptions should not be equal
	assert.False(t, desc1.Equals(desc3))
	assert.False(t, desc3.Equals(desc1))

	// Empty descriptions should be equal
	assert.True(t, emptyDesc1.Equals(emptyDesc2))

	// Empty and non-empty should not be equal
	assert.False(t, emptyDesc1.Equals(desc1))
}

func TestDescription_Equals_WithDifferentValueObjectType(t *testing.T) {
	// Arrange
	description := NewDescription("User Service")

	// Create a different value object type (ComponentName) for comparison
	componentName, err := NewComponentName("User Service")
	assert.NoError(t, err)

	// Act & Assert: Different value object types should not be equal even with same string value
	assert.False(t, description.Equals(componentName))
}

func TestDescription_String(t *testing.T) {
	// Arrange
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
			// Act
			description := NewDescription(tc.input)

			// Assert
			assert.Equal(t, tc.expected, description.String())
		})
	}
}

func TestDescription_LongText(t *testing.T) {
	// Arrange: Long description text
	longText := "This is a very long description that contains multiple sentences. " +
		"It describes the component in great detail, providing context about its purpose, " +
		"functionality, and integration points. The description can span multiple lines " +
		"and include technical details about the component's implementation."

	// Act
	description := NewDescription(longText)

	// Assert: Should handle long text without issues
	assert.Equal(t, longText, description.Value())
	assert.False(t, description.IsEmpty())
}

func TestDescription_SpecialCharacters(t *testing.T) {
	// Arrange: Description with special characters
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
			// Act
			description := NewDescription(tc.input)

			// Assert: Should preserve special characters (after trimming)
			assert.Contains(t, description.Value(), tc.input)
			assert.False(t, description.IsEmpty())
		})
	}
}
