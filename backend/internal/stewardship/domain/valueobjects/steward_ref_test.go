package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStewardRef_AcceptsAnActiveUser(t *testing.T) {
	ref, err := NewStewardRef("user-1", UserActive)
	require.NoError(t, err)
	assert.Equal(t, "user-1", ref.Value())
}

func TestNewStewardRef_RejectsAnEmptyID(t *testing.T) {
	_, err := NewStewardRef("  ", UserActive)
	assert.ErrorIs(t, err, ErrStewardIDRequired)
}

func TestNewStewardRef_RejectsAnUnknownUser(t *testing.T) {
	_, err := NewStewardRef("user-1", UserUnknown)
	assert.ErrorIs(t, err, ErrStewardNotFound)
}

func TestNewStewardRef_RejectsADisabledUser(t *testing.T) {
	_, err := NewStewardRef("user-1", UserDisabled)
	assert.ErrorIs(t, err, ErrStewardDisabled)
}

func TestStewardRef_EqualityIsByUserID(t *testing.T) {
	a, _ := NewStewardRef("user-1", UserActive)
	b, _ := NewStewardRef("user-1", UserActive)
	c, _ := NewStewardRef("user-2", UserActive)
	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
}
