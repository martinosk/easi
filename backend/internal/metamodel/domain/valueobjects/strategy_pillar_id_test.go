package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStrategyPillarIDFromString_Valid(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"

	id, err := NewStrategyPillarIDFromString(validUUID)

	require.NoError(t, err)
	assert.Equal(t, validUUID, id.Value())
}

func TestNewStrategyPillarIDFromString_Invalid(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"not a uuid", "not-a-uuid"},
		{"partial uuid", "550e8400-e29b-41d4"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewStrategyPillarIDFromString(tc.input)

			assert.Error(t, err)
		})
	}
}
