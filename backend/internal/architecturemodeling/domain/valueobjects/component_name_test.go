package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewComponentName_ValidName(t *testing.T) {
	name, err := NewComponentName("User Service")
	assert.NoError(t, err)
	assert.Equal(t, "User Service", name.Value())
}

func TestNewComponentName_EmptyString(t *testing.T) {
	_, err := NewComponentName("")
	assert.Error(t, err)
	assert.Equal(t, ErrComponentNameEmpty, err)
}

func TestNewComponentName_WhitespaceOnly(t *testing.T) {
	_, err := NewComponentName("   ")
	assert.Error(t, err)
	assert.Equal(t, ErrComponentNameEmpty, err)
}

func TestNewComponentName_TrimsWhitespace(t *testing.T) {
	name, err := NewComponentName("  User Service  ")
	assert.NoError(t, err)
	assert.Equal(t, "User Service", name.Value())
}

func TestComponentName_Equals(t *testing.T) {
	name1, _ := NewComponentName("User Service")
	name2, _ := NewComponentName("User Service")
	name3, _ := NewComponentName("Order Service")

	assert.True(t, name1.Equals(name2))
	assert.False(t, name1.Equals(name3))
}
