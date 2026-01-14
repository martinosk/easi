package entities

import (
	"testing"

	"easi/backend/internal/architecturemodeling/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExpert_ValidInput(t *testing.T) {
	expert, err := NewExpert("Alice Smith", "Product Owner", "alice@example.com")

	require.NoError(t, err)
	assert.Equal(t, "Alice Smith", expert.Name().Value())
	assert.Equal(t, "Product Owner", expert.Role().Value())
	assert.Equal(t, "alice@example.com", expert.Contact().Value())
	assert.False(t, expert.AddedAt().IsZero())
}

func TestNewExpert_TrimsWhitespace(t *testing.T) {
	expert, err := NewExpert("  Alice Smith  ", "  Product Owner  ", "  alice@example.com  ")

	require.NoError(t, err)
	assert.Equal(t, "Alice Smith", expert.Name().Value())
	assert.Equal(t, "Product Owner", expert.Role().Value())
	assert.Equal(t, "alice@example.com", expert.Contact().Value())
}

func TestNewExpert_EmptyName(t *testing.T) {
	_, err := NewExpert("", "Product Owner", "alice@example.com")

	require.Error(t, err)
	assert.ErrorIs(t, err, valueobjects.ErrExpertNameEmpty)
}

func TestNewExpert_WhitespaceOnlyName(t *testing.T) {
	_, err := NewExpert("   ", "Product Owner", "alice@example.com")

	require.Error(t, err)
	assert.ErrorIs(t, err, valueobjects.ErrExpertNameEmpty)
}

func TestNewExpert_EmptyRole(t *testing.T) {
	_, err := NewExpert("Alice Smith", "", "alice@example.com")

	require.Error(t, err)
	assert.ErrorIs(t, err, valueobjects.ErrExpertRoleEmpty)
}

func TestNewExpert_WhitespaceOnlyRole(t *testing.T) {
	_, err := NewExpert("Alice Smith", "   ", "alice@example.com")

	require.Error(t, err)
	assert.ErrorIs(t, err, valueobjects.ErrExpertRoleEmpty)
}

func TestNewExpert_EmptyContact(t *testing.T) {
	_, err := NewExpert("Alice Smith", "Product Owner", "")

	require.Error(t, err)
	assert.ErrorIs(t, err, valueobjects.ErrContactInfoEmpty)
}

func TestNewExpert_WhitespaceOnlyContact(t *testing.T) {
	_, err := NewExpert("Alice Smith", "Product Owner", "   ")

	require.Error(t, err)
	assert.ErrorIs(t, err, valueobjects.ErrContactInfoEmpty)
}
