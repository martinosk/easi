package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRealizationRolesID(t *testing.T) {
	id := NewRealizationRolesID()
	assert.NotEmpty(t, id.Value())
}

func TestNewRealizationRolesIDFromString_Valid(t *testing.T) {
	id := NewRealizationRolesID()
	parsed, err := NewRealizationRolesIDFromString(id.Value())
	require.NoError(t, err)
	assert.Equal(t, id.Value(), parsed.Value())
}

func TestNewRealizationRolesIDFromString_Empty(t *testing.T) {
	_, err := NewRealizationRolesIDFromString("")
	assert.Error(t, err)
}

func TestNewRealizationRolesIDFromString_Invalid(t *testing.T) {
	_, err := NewRealizationRolesIDFromString("not-a-uuid")
	assert.Error(t, err)
}

func TestRealizationRolesID_Equals(t *testing.T) {
	id := NewRealizationRolesID()
	parsed, _ := NewRealizationRolesIDFromString(id.Value())
	assert.True(t, id.Equals(parsed))
}
