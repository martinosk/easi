package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJourneyProgress_Valid_Rule7(t *testing.T) {
	for _, v := range []int{0, 1, 60, 99, 100} {
		p, err := NewJourneyProgress(v)
		require.NoError(t, err)
		assert.Equal(t, v, p.Value())
	}
}

func TestNewJourneyProgress_OutOfRange(t *testing.T) {
	_, err := NewJourneyProgress(-1)
	assert.ErrorIs(t, err, ErrInvalidJourneyProgress)

	_, err = NewJourneyProgress(101)
	assert.ErrorIs(t, err, ErrInvalidJourneyProgress)
}

func TestJourneyProgress_Equals(t *testing.T) {
	a, _ := NewJourneyProgress(60)
	b, _ := NewJourneyProgress(60)
	c, _ := NewJourneyProgress(40)
	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
}
