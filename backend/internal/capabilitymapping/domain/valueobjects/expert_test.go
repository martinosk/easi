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
	assert.Equal(t, "Alice Smith", expert.Name())
	assert.Equal(t, "Product Owner", expert.Role())
	assert.Equal(t, "alice@example.com", expert.Contact())
	assert.Equal(t, addedAt, expert.AddedAt())
}

func TestNewExpert_TrimsWhitespace(t *testing.T) {
	addedAt := time.Now().UTC()
	expert, err := NewExpert("  Alice Smith  ", "  Product Owner  ", "  alice@example.com  ", addedAt)

	require.NoError(t, err)
	assert.Equal(t, "Alice Smith", expert.Name())
	assert.Equal(t, "Product Owner", expert.Role())
	assert.Equal(t, "alice@example.com", expert.Contact())
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
	assert.ErrorIs(t, err, ErrExpertContactEmpty)
}

func TestNewExpert_WhitespaceOnlyContact(t *testing.T) {
	_, err := NewExpert("Alice Smith", "Product Owner", "   ", time.Now().UTC())

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrExpertContactEmpty)
}

func TestMustNewExpert_ValidInput(t *testing.T) {
	addedAt := time.Now().UTC()
	expert := MustNewExpert("Alice Smith", "Product Owner", "alice@example.com", addedAt)

	assert.Equal(t, "Alice Smith", expert.Name())
	assert.Equal(t, "Product Owner", expert.Role())
	assert.Equal(t, "alice@example.com", expert.Contact())
}

func TestMustNewExpert_PanicsOnInvalidInput(t *testing.T) {
	assert.Panics(t, func() {
		MustNewExpert("", "Product Owner", "alice@example.com", time.Now().UTC())
	})
}
