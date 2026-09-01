package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOwnershipState_ValidValues(t *testing.T) {
	for _, value := range []string{
		OwnershipStateUnknown,
		OwnershipStateNominated,
		OwnershipStateOwned,
		OwnershipStateManaged,
	} {
		state, err := NewOwnershipState(value)
		require.NoError(t, err)
		assert.Equal(t, value, state.String())
	}
}

func TestNewOwnershipState_InvalidValue(t *testing.T) {
	_, err := NewOwnershipState("orphaned")
	assert.ErrorIs(t, err, ErrInvalidOwnershipState)
}

func TestUnknownOwnershipState(t *testing.T) {
	state := UnknownOwnershipState()
	assert.Equal(t, OwnershipStateUnknown, state.String())
	assert.True(t, state.IsUnknown())
}

func TestOwnershipState_Equals(t *testing.T) {
	a, _ := NewOwnershipState(OwnershipStateOwned)
	b, _ := NewOwnershipState(OwnershipStateOwned)
	c, _ := NewOwnershipState(OwnershipStateManaged)

	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
}
