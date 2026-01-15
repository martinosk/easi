package valueobjects

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExpert_ValidInput(t *testing.T) {
	addedAt := time.Now().UTC()
	expert, err := NewExpert("Alice Smith", "Product Owner", "alice@example.com", addedAt)

	require.NoError(t, err)
	assert.Equal(t, "Alice Smith", expert.Name().Value())
	assert.Equal(t, "Product Owner", expert.Role().Value())
	assert.Equal(t, "alice@example.com", expert.Contact().Value())
	assert.Equal(t, addedAt, expert.AddedAt())
}

func TestNewExpert_TrimsWhitespace(t *testing.T) {
	addedAt := time.Now().UTC()
	expert, err := NewExpert("  Alice Smith  ", "  Product Owner  ", "  alice@example.com  ", addedAt)

	require.NoError(t, err)
	assert.Equal(t, "Alice Smith", expert.Name().Value())
	assert.Equal(t, "Product Owner", expert.Role().Value())
	assert.Equal(t, "alice@example.com", expert.Contact().Value())
}

func TestNewExpert_EmptyName(t *testing.T) {
	_, err := NewExpert("", "Product Owner", "alice@example.com", time.Now().UTC())

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrExpertNameEmpty)
}

func TestNewExpert_WhitespaceOnlyName(t *testing.T) {
	_, err := NewExpert("   ", "Product Owner", "alice@example.com", time.Now().UTC())

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrExpertNameEmpty)
}

func TestNewExpert_EmptyRole(t *testing.T) {
	_, err := NewExpert("Alice Smith", "", "alice@example.com", time.Now().UTC())

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrExpertRoleEmpty)
}

func TestNewExpert_WhitespaceOnlyRole(t *testing.T) {
	_, err := NewExpert("Alice Smith", "   ", "alice@example.com", time.Now().UTC())

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrExpertRoleEmpty)
}

func TestNewExpert_EmptyContact(t *testing.T) {
	_, err := NewExpert("Alice Smith", "Product Owner", "", time.Now().UTC())

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrContactInfoEmpty)
}

func TestNewExpert_WhitespaceOnlyContact(t *testing.T) {
	_, err := NewExpert("Alice Smith", "Product Owner", "   ", time.Now().UTC())

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrContactInfoEmpty)
}

func TestMustNewExpert_ValidInput(t *testing.T) {
	addedAt := time.Now().UTC()
	expert := MustNewExpert("Alice Smith", "Product Owner", "alice@example.com", addedAt)

	assert.Equal(t, "Alice Smith", expert.Name().Value())
	assert.Equal(t, "Product Owner", expert.Role().Value())
	assert.Equal(t, "alice@example.com", expert.Contact().Value())
}

func TestMustNewExpert_PanicsOnInvalidInput(t *testing.T) {
	assert.Panics(t, func() {
		MustNewExpert("", "Product Owner", "alice@example.com", time.Now().UTC())
	})
}

func TestExpert_Equals(t *testing.T) {
	addedAt1 := time.Now().UTC()
	addedAt2 := addedAt1.Add(time.Hour)

	expert1, _ := NewExpert("Alice Smith", "Product Owner", "alice@example.com", addedAt1)
	expert2, _ := NewExpert("Alice Smith", "Product Owner", "alice@example.com", addedAt2)
	expert3, _ := NewExpert("Bob Jones", "Product Owner", "alice@example.com", addedAt1)

	assert.True(t, expert1.Equals(expert2))
	assert.False(t, expert1.Equals(expert3))
}
