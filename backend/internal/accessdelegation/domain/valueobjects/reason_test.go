package valueobjects

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReason_EmptyIsValid(t *testing.T) {
	r, err := NewReason("")
	require.NoError(t, err)
	assert.Equal(t, "", r.Value())
}

func TestNewReason_AtMaxLength_IsValid(t *testing.T) {
	input := strings.Repeat("a", MaxReasonLength)
	r, err := NewReason(input)
	require.NoError(t, err)
	assert.Equal(t, input, r.Value())
}

func TestNewReason_ExceedsMaxLength_ReturnsError(t *testing.T) {
	input := strings.Repeat("a", MaxReasonLength+1)
	_, err := NewReason(input)
	assert.Equal(t, ErrReasonTooLong, err)
}
