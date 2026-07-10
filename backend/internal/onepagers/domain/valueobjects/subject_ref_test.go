package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSubjectRef_AcceptsSubjectTypeAndID(t *testing.T) {
	ref, err := NewSubjectRef("application", "  app-123  ")
	require.NoError(t, err)
	assert.Equal(t, "application", ref.SubjectType().Value())
	assert.Equal(t, "app-123", ref.SubjectID())
}

func TestNewSubjectRef_RejectsInvalidInput(t *testing.T) {
	_, err := NewSubjectRef("starship", "app-123")
	assert.ErrorIs(t, err, ErrInvalidSubjectType)

	_, err = NewSubjectRef("application", "   ")
	assert.ErrorIs(t, err, ErrSubjectIDEmpty)
}

func TestSubjectRef_Equals(t *testing.T) {
	a, err := NewSubjectRef("application", "app-123")
	require.NoError(t, err)
	b, err := NewSubjectRef("application", "app-123")
	require.NoError(t, err)
	c, err := NewSubjectRef("vendor", "app-123")
	require.NoError(t, err)

	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
}
