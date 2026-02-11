package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValueStreamID_GeneratesUniqueIDs(t *testing.T) {
	id1 := NewValueStreamID()
	id2 := NewValueStreamID()

	assert.NotEmpty(t, id1.Value())
	assert.NotEmpty(t, id2.Value())
	assert.NotEqual(t, id1.Value(), id2.Value())
}

func TestNewValueStreamIDFromString_Valid(t *testing.T) {
	original := NewValueStreamID()
	restored, err := NewValueStreamIDFromString(original.Value())
	require.NoError(t, err)
	assert.Equal(t, original.Value(), restored.Value())
}

func TestNewValueStreamIDFromString_Invalid(t *testing.T) {
	_, err := NewValueStreamIDFromString("")
	assert.Error(t, err)

	_, err = NewValueStreamIDFromString("not-a-uuid")
	assert.Error(t, err)
}

func TestValueStreamID_Equals(t *testing.T) {
	id1 := NewValueStreamID()
	id2, _ := NewValueStreamIDFromString(id1.Value())
	id3 := NewValueStreamID()

	assert.True(t, id1.Equals(id2))
	assert.False(t, id1.Equals(id3))
}
