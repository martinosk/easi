package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDirectionID(t *testing.T) {
	id := NewDirectionID()
	assert.NotEmpty(t, id.Value())
}

func TestNewDirectionIDFromString_Valid(t *testing.T) {
	id := NewDirectionID()
	parsed, err := NewDirectionIDFromString(id.Value())
	require.NoError(t, err)
	assert.Equal(t, id.Value(), parsed.Value())
}

func TestNewDirectionIDFromString_Empty(t *testing.T) {
	_, err := NewDirectionIDFromString("")
	assert.Error(t, err)
}

func TestNewDirectionIDFromString_Invalid(t *testing.T) {
	_, err := NewDirectionIDFromString("not-a-uuid")
	assert.Error(t, err)
}

func TestDirectionID_Equals(t *testing.T) {
	id := NewDirectionID()
	parsed, _ := NewDirectionIDFromString(id.Value())
	assert.True(t, id.Equals(parsed))
}
