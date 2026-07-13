package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJourneyStatus_AllValid(t *testing.T) {
	for _, v := range []string{"planned", "in-flight", "done", "abandoned"} {
		s, err := NewJourneyStatus(v)
		require.NoError(t, err)
		assert.Equal(t, v, s.Value())
	}
}

func TestNewJourneyStatus_Invalid(t *testing.T) {
	_, err := NewJourneyStatus("pending")
	assert.ErrorIs(t, err, ErrInvalidJourneyStatus)
}

func TestJourneyStatus_IsActive_Rule6(t *testing.T) {
	cases := map[string]bool{
		"planned":   true,
		"in-flight": true,
		"done":      false,
		"abandoned": false,
	}
	for v, active := range cases {
		s, _ := NewJourneyStatus(v)
		assert.Equal(t, active, s.IsActive(), v)
		assert.Equal(t, !active, s.IsTerminal(), v)
	}
}

func TestJourneyStatus_CanStart_OnlyFromPlanned(t *testing.T) {
	cases := map[string]bool{
		"planned":   true,
		"in-flight": false,
		"done":      false,
		"abandoned": false,
	}
	for v, can := range cases {
		s, _ := NewJourneyStatus(v)
		assert.Equal(t, can, s.CanStart(), v)
	}
}

func TestJourneyStatus_CanComplete_OnlyFromInFlight(t *testing.T) {
	cases := map[string]bool{
		"planned":   false,
		"in-flight": true,
		"done":      false,
		"abandoned": false,
	}
	for v, can := range cases {
		s, _ := NewJourneyStatus(v)
		assert.Equal(t, can, s.CanComplete(), v)
	}
}

func TestJourneyStatus_CanAbandon_FromPlannedOrInFlight(t *testing.T) {
	cases := map[string]bool{
		"planned":   true,
		"in-flight": true,
		"done":      false,
		"abandoned": false,
	}
	for v, can := range cases {
		s, _ := NewJourneyStatus(v)
		assert.Equal(t, can, s.CanAbandon(), v)
	}
}

func TestJourneyStatus_Equals(t *testing.T) {
	a, _ := NewJourneyStatus("planned")
	b, _ := NewJourneyStatus("planned")
	c, _ := NewJourneyStatus("done")
	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
}
