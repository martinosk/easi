package valueobjects

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewResultingCapabilityName_Valid_Rule5(t *testing.T) {
	n, err := NewResultingCapabilityName("  Freight invoicing  ")
	require.NoError(t, err)
	assert.Equal(t, "Freight invoicing", n.Value())
}

func TestNewResultingCapabilityName_Empty_Rejected(t *testing.T) {
	_, err := NewResultingCapabilityName("   ")
	assert.ErrorIs(t, err, ErrResultingCapabilityNameRequired)
}

func TestNewResultingCapabilityName_TooLong_Rejected(t *testing.T) {
	_, err := NewResultingCapabilityName(strings.Repeat("a", 201))
	assert.ErrorIs(t, err, ErrResultingCapabilityNameTooLong)
}

func TestNewResultingCapabilityName_MaxLength_Accepted(t *testing.T) {
	_, err := NewResultingCapabilityName(strings.Repeat("a", 200))
	assert.NoError(t, err)
}
