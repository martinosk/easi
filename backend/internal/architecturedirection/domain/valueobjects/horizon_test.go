package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHorizon_Valid(t *testing.T) {
	for _, v := range []string{"now", "next", "later"} {
		h, err := NewHorizon(v)
		require.NoError(t, err)
		assert.Equal(t, v, h.Value())
	}
}

func TestNewHorizon_Invalid(t *testing.T) {
	_, err := NewHorizon("eventually")
	assert.ErrorIs(t, err, ErrInvalidHorizon)
}

func TestNewHorizon_Empty(t *testing.T) {
	_, err := NewHorizon("")
	assert.ErrorIs(t, err, ErrInvalidHorizon)
}
