package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewViewName_ValidName(t *testing.T) {
	name, err := NewViewName("User Management View")
	assert.NoError(t, err)
	assert.Equal(t, "User Management View", name.Value())
}

func TestNewViewName_EmptyString(t *testing.T) {
	_, err := NewViewName("")
	assert.Error(t, err)
	assert.Equal(t, ErrViewNameEmpty, err)
}

func TestNewViewName_WhitespaceOnly(t *testing.T) {
	_, err := NewViewName("   ")
	assert.Error(t, err)
	assert.Equal(t, ErrViewNameEmpty, err)
}

func TestNewViewName_TrimsWhitespace(t *testing.T) {
	name, err := NewViewName("  Payment Flow  ")
	assert.NoError(t, err)
	assert.Equal(t, "Payment Flow", name.Value())
}

func TestViewName_Equals(t *testing.T) {
	name1, _ := NewViewName("Dashboard")
	name2, _ := NewViewName("Dashboard")
	name3, _ := NewViewName("Reports")

	assert.True(t, name1.Equals(name2))
	assert.False(t, name1.Equals(name3))
}

func TestViewName_String(t *testing.T) {
	name, _ := NewViewName("Main View")
	assert.Equal(t, "Main View", name.String())
}

func TestNewViewName_MaxLength(t *testing.T) {
	// Test exactly 100 characters (should pass)
	validName := "This is a view name that is exactly one hundred characters long to test the maximum length validatio"
	assert.Len(t, validName, 100, "Test string should be exactly 100 characters")

	name, err := NewViewName(validName)
	assert.NoError(t, err)
	assert.Equal(t, validName, name.Value())
}

func TestNewViewName_ExceedsMaxLength(t *testing.T) {
	// Test 101 characters (should fail)
	tooLongName := "This is a view name that is one hundred and one characters long and should fail the validation tests!"
	assert.Len(t, tooLongName, 101, "Test string should be exactly 101 characters")

	_, err := NewViewName(tooLongName)
	assert.Error(t, err)
	assert.Equal(t, ErrViewNameTooLong, err)
}

func TestNewViewName_SingleCharacter(t *testing.T) {
	// Test minimum length (1 character)
	name, err := NewViewName("A")
	assert.NoError(t, err)
	assert.Equal(t, "A", name.Value())
}
