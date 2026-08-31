package valueobjects

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRationale_TrimsWhitespace(t *testing.T) {
	rationale, err := NewRationale("  some rationale  ")
	require.NoError(t, err)
	assert.Equal(t, "some rationale", rationale.Value())
}

func TestNewRationale_AcceptsEmptyRationale(t *testing.T) {
	rationale, err := NewRationale("")
	require.NoError(t, err)
	assert.Empty(t, rationale.Value())
	assert.True(t, rationale.IsEmpty())
}

func TestNewRationale_AcceptsMaxLength(t *testing.T) {
	maxLengthRationale := strings.Repeat("a", MaxRationaleLength)
	rationale, err := NewRationale(maxLengthRationale)
	require.NoError(t, err)
	assert.Equal(t, MaxRationaleLength, len(rationale.Value()))
}

func TestNewRationale_RejectsTooLong(t *testing.T) {
	tooLongRationale := strings.Repeat("a", MaxRationaleLength+1)
	_, err := NewRationale(tooLongRationale)
	assert.ErrorIs(t, err, ErrRationaleTooLong)
}

func TestEmptyRationale(t *testing.T) {
	rationale := EmptyRationale()
	assert.Empty(t, rationale.Value())
	assert.True(t, rationale.IsEmpty())
}

func TestRationale_IsEmpty(t *testing.T) {
	emptyRationale, _ := NewRationale("")
	assert.True(t, emptyRationale.IsEmpty())

	nonEmptyRationale, _ := NewRationale("test")
	assert.False(t, nonEmptyRationale.IsEmpty())
}

func TestRationale_Equals(t *testing.T) {
	rat1, _ := NewRationale("test")
	rat2, _ := NewRationale("test")
	rat3, _ := NewRationale("other")

	assert.True(t, rat1.Equals(rat2))
	assert.False(t, rat1.Equals(rat3))
}
