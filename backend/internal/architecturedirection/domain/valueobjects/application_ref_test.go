package valueobjects

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewApplicationRef_Valid(t *testing.T) {
	id := uuid.New().String()
	ref, err := NewApplicationRef(id)
	require.NoError(t, err)
	assert.Equal(t, id, ref.Value())
}

func TestNewApplicationRef_Invalid(t *testing.T) {
	_, err := NewApplicationRef("not-uuid")
	assert.Error(t, err)
}

func TestApplicationRef_Equals(t *testing.T) {
	id := uuid.New().String()
	a, _ := NewApplicationRef(id)
	b, _ := NewApplicationRef(id)
	c, _ := NewApplicationRef(uuid.New().String())
	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
}
