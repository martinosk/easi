package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOwnerReference_User(t *testing.T) {
	ref, err := NewOwnerReference(OwnerKindUser, "user-123")
	require.NoError(t, err)
	assert.Equal(t, OwnerKindUser, ref.Kind())
	assert.Equal(t, "user-123", ref.ID())
	assert.True(t, ref.IsUser())
	assert.False(t, ref.IsTeam())
}

func TestNewOwnerReference_Team(t *testing.T) {
	ref, err := NewOwnerReference(OwnerKindTeam, "team-456")
	require.NoError(t, err)
	assert.True(t, ref.IsTeam())
	assert.False(t, ref.IsUser())
}

func TestNewOwnerReference_InvalidKind(t *testing.T) {
	_, err := NewOwnerReference("department", "dep-1")
	assert.ErrorIs(t, err, ErrInvalidOwnerKind)
}

func TestNewOwnerReference_EmptyID(t *testing.T) {
	_, err := NewOwnerReference(OwnerKindUser, "")
	assert.ErrorIs(t, err, ErrEmptyOwnerID)
}

func TestOwnerReference_Equals(t *testing.T) {
	a, _ := NewOwnerReference(OwnerKindUser, "user-123")
	b, _ := NewOwnerReference(OwnerKindUser, "user-123")
	c, _ := NewOwnerReference(OwnerKindTeam, "user-123")

	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
}
