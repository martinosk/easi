package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStandardApplicationID(t *testing.T) {
	id := NewStandardApplicationID()
	assert.NotEmpty(t, id.Value())
}

func TestNewStandardApplicationIDFromString_Valid(t *testing.T) {
	id := NewStandardApplicationID()
	parsed, err := NewStandardApplicationIDFromString(id.Value())
	require.NoError(t, err)
	assert.Equal(t, id.Value(), parsed.Value())
}

func TestNewStandardApplicationIDFromString_Empty(t *testing.T) {
	_, err := NewStandardApplicationIDFromString("")
	assert.Error(t, err)
}

func TestNewStandardApplicationIDFromString_Invalid(t *testing.T) {
	_, err := NewStandardApplicationIDFromString("not-a-uuid")
	assert.Error(t, err)
}

func TestStandardApplicationID_Equals(t *testing.T) {
	id := NewStandardApplicationID()
	parsed, _ := NewStandardApplicationIDFromString(id.Value())
	assert.True(t, id.Equals(parsed))
}
