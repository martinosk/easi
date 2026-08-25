package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEAOwner_AcceptsEmpty(t *testing.T) {
	ref, err := NewEAOwner("")
	require.NoError(t, err)
	assert.True(t, ref.IsEmpty())
	assert.Equal(t, "", ref.Value())
}

func TestNewEAOwner_AcceptsWhitespaceAsEmpty(t *testing.T) {
	ref, err := NewEAOwner("   ")
	require.NoError(t, err)
	assert.True(t, ref.IsEmpty())
}

func TestNewEAOwner_AcceptsUserID(t *testing.T) {
	ref, err := NewEAOwner("2ec46b70-63b3-4d6d-92f0-1d385f9d4c4b")
	require.NoError(t, err)
	assert.False(t, ref.IsEmpty())
	assert.Equal(t, "2ec46b70-63b3-4d6d-92f0-1d385f9d4c4b", ref.Value())
}

func TestNewEAOwner_TrimsWhitespaceAroundUserID(t *testing.T) {
	ref, err := NewEAOwner("  2ec46b70-63b3-4d6d-92f0-1d385f9d4c4b  ")
	require.NoError(t, err)
	assert.Equal(t, "2ec46b70-63b3-4d6d-92f0-1d385f9d4c4b", ref.Value())
}

func TestNewEAOwner_RejectsFreeText(t *testing.T) {
	_, err := NewEAOwner("Alice Smith")
	assert.ErrorIs(t, err, ErrEAOwnerNotUser)
}

func TestEAOwnerFromHistory_AcceptsAnyValue(t *testing.T) {
	ref := EAOwnerFromHistory("Alice Smith")
	assert.Equal(t, "Alice Smith", ref.Value())
	assert.False(t, ref.IsEmpty())
}

func TestEAOwner_Equals(t *testing.T) {
	a := EAOwnerFromHistory("x")
	b := EAOwnerFromHistory("x")
	c := EAOwnerFromHistory("y")
	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
	assert.False(t, a.Equals(NewOwner("x")))
}
