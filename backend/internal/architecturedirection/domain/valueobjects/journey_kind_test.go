package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJourneyKind_AllValid(t *testing.T) {
	for _, v := range []string{"migration", "consolidation", "carve-out", "move"} {
		k, err := NewJourneyKind(v)
		require.NoError(t, err)
		assert.Equal(t, v, k.Value())
	}
}

func TestNewJourneyKind_Invalid(t *testing.T) {
	_, err := NewJourneyKind("relocation")
	assert.ErrorIs(t, err, ErrInvalidJourneyKind)
}

func TestJourneyKind_IsMove(t *testing.T) {
	move, _ := NewJourneyKind("move")
	assert.True(t, move.IsMove())

	migration, _ := NewJourneyKind("migration")
	assert.False(t, migration.IsMove())
}

func TestJourneyKind_ValidateSourceCount_Rule3(t *testing.T) {
	cases := []struct {
		kind  string
		count int
		valid bool
	}{
		{"migration", 0, false},
		{"migration", 1, true},
		{"migration", 3, true},
		{"consolidation", 0, false},
		{"consolidation", 1, false},
		{"consolidation", 2, true},
		{"consolidation", 5, true},
		{"carve-out", 0, false},
		{"carve-out", 1, true},
		{"carve-out", 2, false},
		{"move", 0, true},
		{"move", 1, true},
		{"move", 4, true},
	}
	for _, c := range cases {
		t.Run(c.kind, func(t *testing.T) {
			k, err := NewJourneyKind(c.kind)
			require.NoError(t, err)
			err = k.ValidateSourceCount(c.count)
			if c.valid {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, ErrInvalidSourceApplicationCount)
			}
		})
	}
}

func TestJourneyKind_Equals(t *testing.T) {
	a, _ := NewJourneyKind("migration")
	b, _ := NewJourneyKind("migration")
	c, _ := NewJourneyKind("move")
	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
}
