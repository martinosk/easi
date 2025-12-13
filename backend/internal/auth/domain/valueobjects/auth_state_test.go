package valueobjects

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuthState_GeneratesValidState(t *testing.T) {
	state := NewAuthState()

	value := state.Value()
	assert.NotEmpty(t, value, "state should not be empty")

	urlSafe := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	assert.True(t, urlSafe.MatchString(value), "state should be URL-safe")
}

func TestNewAuthState_GeneratesUniqueValues(t *testing.T) {
	state1 := NewAuthState()
	state2 := NewAuthState()

	assert.NotEqual(t, state1.Value(), state2.Value(), "each state should be unique")
}

func TestAuthStateFromValue_ValidState(t *testing.T) {
	original := NewAuthState()

	restored, err := AuthStateFromValue(original.Value())
	require.NoError(t, err)

	assert.Equal(t, original.Value(), restored.Value())
}

func TestAuthStateFromValue_EmptyState(t *testing.T) {
	_, err := AuthStateFromValue("")
	assert.Error(t, err)
}

func TestAuthState_Equals(t *testing.T) {
	state1 := NewAuthState()

	state2, err := AuthStateFromValue(state1.Value())
	require.NoError(t, err)

	state3 := NewAuthState()

	assert.True(t, state1.Equals(state2), "same value states should be equal")
	assert.False(t, state1.Equals(state3), "different states should not be equal")
}
