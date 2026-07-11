package valueobjects

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFieldID_GeneratesUniqueUUIDs(t *testing.T) {
	a := NewFieldID()
	b := NewFieldID()
	_, err := uuid.Parse(a.Value())
	require.NoError(t, err)
	assert.NotEqual(t, a.Value(), b.Value())
}

func TestNewFieldIDFromString_AcceptsValidUUID(t *testing.T) {
	raw := uuid.New().String()
	id, err := NewFieldIDFromString(raw)
	require.NoError(t, err)
	assert.Equal(t, raw, id.Value())
}

func TestNewFieldIDFromString_RejectsInvalidUUID(t *testing.T) {
	_, err := NewFieldIDFromString("not-a-uuid")
	assert.ErrorIs(t, err, ErrInvalidFieldID)
}

func TestNewOptionID_GeneratesUniqueUUIDs(t *testing.T) {
	a := NewOptionID()
	b := NewOptionID()
	_, err := uuid.Parse(a.Value())
	require.NoError(t, err)
	assert.NotEqual(t, a.Value(), b.Value())
}

func TestNewOptionIDFromString_RejectsInvalidUUID(t *testing.T) {
	_, err := NewOptionIDFromString("nope")
	assert.ErrorIs(t, err, ErrInvalidOptionID)
}

func TestFieldID_Equals(t *testing.T) {
	raw := uuid.New().String()
	a, _ := NewFieldIDFromString(raw)
	b, _ := NewFieldIDFromString(raw)
	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(NewFieldID()))
}
