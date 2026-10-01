package home

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusCounts_AddCountsEveryCapabilityStatus(t *testing.T) {
	var counts StatusCounts
	require.NoError(t, counts.Add("Active", 10))
	require.NoError(t, counts.Add("Planned", 1))
	require.NoError(t, counts.Add("Deprecated", 2))

	assert.Equal(t, StatusCounts{Active: 10, Planned: 1, Deprecated: 2}, counts)
	assert.Equal(t, 13, counts.total())
}

func TestStatusCounts_AddRefusesAStatusItDoesNotKnow(t *testing.T) {
	var counts StatusCounts

	err := counts.Add("Retired", 3)

	assert.ErrorIs(t, err, ErrUnknownCapabilityStatus)
	assert.Zero(t, counts.total())
}
