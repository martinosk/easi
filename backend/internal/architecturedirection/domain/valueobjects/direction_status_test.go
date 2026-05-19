package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDirectionStatus_AllValid(t *testing.T) {
	for _, v := range []string{"draft", "proposed", "agreed", "rejected"} {
		s, err := NewDirectionStatus(v)
		require.NoError(t, err)
		assert.Equal(t, v, s.Value())
	}
}

func TestNewDirectionStatus_Invalid(t *testing.T) {
	_, err := NewDirectionStatus("approved")
	assert.ErrorIs(t, err, ErrInvalidDirectionStatus)
}

func TestDirectionStatus_IsActive(t *testing.T) {
	cases := map[string]bool{
		"draft":    true,
		"proposed": true,
		"agreed":   true,
		"rejected": false,
	}
	for v, active := range cases {
		s, _ := NewDirectionStatus(v)
		assert.Equal(t, active, s.IsActive(), v)
	}
}

func TestDirectionStatus_IsTerminal(t *testing.T) {
	rejected, _ := NewDirectionStatus("rejected")
	assert.True(t, rejected.IsTerminal())

	draft, _ := NewDirectionStatus("draft")
	assert.False(t, draft.IsTerminal())
}

func TestDirectionStatus_CanAdvanceTo(t *testing.T) {
	cases := []struct {
		from, to string
		ok       bool
	}{
		{"draft", "proposed", true},
		{"proposed", "agreed", true},
		{"draft", "agreed", false},   // skip-step forbidden
		{"agreed", "proposed", false}, // forward-only
		{"agreed", "agreed", false},
		{"rejected", "proposed", false},
	}
	for _, c := range cases {
		t.Run(c.from+"->"+c.to, func(t *testing.T) {
			from, _ := NewDirectionStatus(c.from)
			to, _ := NewDirectionStatus(c.to)
			assert.Equal(t, c.ok, from.CanAdvanceTo(to))
		})
	}
}

func TestDirectionStatus_CanReject(t *testing.T) {
	cases := map[string]bool{
		"draft":    true,
		"proposed": true,
		"agreed":   true,
		"rejected": false,
	}
	for v, can := range cases {
		s, _ := NewDirectionStatus(v)
		assert.Equal(t, can, s.CanReject(), v)
	}
}
