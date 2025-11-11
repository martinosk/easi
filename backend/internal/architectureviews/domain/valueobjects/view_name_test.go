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
