package valueobjects

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNarrative_Valid(t *testing.T) {
	n, err := NewNarrative("We consolidate payroll services into one.")
	require.NoError(t, err)
	assert.Equal(t, "We consolidate payroll services into one.", n.Value())
	assert.False(t, n.IsEmpty())
}

func TestNewNarrative_Trimmed(t *testing.T) {
	n, err := NewNarrative("  hello world  ")
	require.NoError(t, err)
	assert.Equal(t, "hello world", n.Value())
}

func TestNewNarrative_Empty(t *testing.T) {
	n, err := NewNarrative("")
	require.NoError(t, err)
	assert.True(t, n.IsEmpty())
}

func TestNewNarrative_WhitespaceOnly(t *testing.T) {
	n, err := NewNarrative("   \t\n  ")
	require.NoError(t, err)
	assert.True(t, n.IsEmpty())
}

func TestNewNarrative_TooLong(t *testing.T) {
	tooLong := strings.Repeat("a", MaxNarrativeLength+1)
	_, err := NewNarrative(tooLong)
	assert.ErrorIs(t, err, ErrNarrativeTooLong)
}

func TestNewNarrative_MaxLength(t *testing.T) {
	atMax := strings.Repeat("a", MaxNarrativeLength)
	n, err := NewNarrative(atMax)
	require.NoError(t, err)
	assert.Equal(t, MaxNarrativeLength, len(n.Value()))
}
