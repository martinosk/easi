package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDirectionType_Consolidate(t *testing.T) {
	dt, err := NewDirectionType("consolidate")
	require.NoError(t, err)
	assert.Equal(t, "consolidate", dt.Value())
	assert.True(t, dt.IsConsolidate())
	assert.False(t, dt.IsDecompose())
	assert.False(t, dt.IsStay())
}

func TestNewDirectionType_Decompose(t *testing.T) {
	dt, err := NewDirectionType("decompose")
	require.NoError(t, err)
	assert.True(t, dt.IsDecompose())
}

func TestNewDirectionType_Stay(t *testing.T) {
	dt, err := NewDirectionType("stay")
	require.NoError(t, err)
	assert.True(t, dt.IsStay())
}

func TestNewDirectionType_Invalid(t *testing.T) {
	_, err := NewDirectionType("expand")
	assert.ErrorIs(t, err, ErrInvalidDirectionType)
}

func TestNewDirectionType_Empty(t *testing.T) {
	_, err := NewDirectionType("")
	assert.ErrorIs(t, err, ErrInvalidDirectionType)
}

func TestDirectionType_RequiredSourceCount(t *testing.T) {
	cases := []struct {
		typ           string
		minSources    int
		exactlyOne    bool
		exactlySources int
	}{
		{"consolidate", 2, false, 0},
		{"decompose", 1, true, 1},
		{"stay", 1, true, 1},
	}
	for _, c := range cases {
		t.Run(c.typ, func(t *testing.T) {
			dt, err := NewDirectionType(c.typ)
			require.NoError(t, err)
			assert.Equal(t, c.exactlyOne, dt.RequiresExactlyOneSource())
			if c.exactlyOne {
				assert.Equal(t, c.exactlySources, dt.ExactSourceCount())
			} else {
				assert.Equal(t, c.minSources, dt.MinSourceCount())
			}
		})
	}
}

func TestDirectionType_PlacementCardinality(t *testing.T) {
	consolidate, _ := NewDirectionType("consolidate")
	decompose, _ := NewDirectionType("decompose")
	stay, _ := NewDirectionType("stay")

	assert.False(t, consolidate.IsValidPlacementCount(0))
	assert.True(t, consolidate.IsValidPlacementCount(1))
	assert.False(t, consolidate.IsValidPlacementCount(2))

	assert.False(t, decompose.IsValidPlacementCount(0))
	assert.True(t, decompose.IsValidPlacementCount(1))
	assert.True(t, decompose.IsValidPlacementCount(2))

	assert.True(t, stay.IsValidPlacementCount(0))
	assert.False(t, stay.IsValidPlacementCount(1))
}
