package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGrantor_EmptyID_ReturnsError(t *testing.T) {
	_, err := NewGrantor("", "test@example.com")
	assert.Equal(t, ErrGrantorIDEmpty, err)
}

func TestNewGrantor_EmptyEmail_ReturnsError(t *testing.T) {
	_, err := NewGrantor("user-id", "")
	assert.Equal(t, ErrGrantorEmailEmpty, err)
}

func TestNewGrantor_ValidInputs_ReturnsCorrectAccessors(t *testing.T) {
	g, err := NewGrantor("user-id", "user@example.com")
	require.NoError(t, err)
	assert.Equal(t, "user-id", g.ID())
	assert.Equal(t, "user@example.com", g.Email())
}

func TestGrantor_Equals_SameValues_ReturnsTrue(t *testing.T) {
	g1, _ := NewGrantor("user-id", "user@example.com")
	g2, _ := NewGrantor("user-id", "user@example.com")
	assert.True(t, g1.Equals(g2))
}

func TestGrantor_Equals_DifferentID_ReturnsFalse(t *testing.T) {
	g1, _ := NewGrantor("user-1", "user@example.com")
	g2, _ := NewGrantor("user-2", "user@example.com")
	assert.False(t, g1.Equals(g2))
}

func TestGrantor_Equals_DifferentEmail_ReturnsFalse(t *testing.T) {
	g1, _ := NewGrantor("user-id", "a@example.com")
	g2, _ := NewGrantor("user-id", "b@example.com")
	assert.False(t, g1.Equals(g2))
}

func TestGrantor_Equals_DifferentValueObjectType_ReturnsFalse(t *testing.T) {
	g, _ := NewGrantor("user-id", "user@example.com")
	assert.False(t, g.Equals(GrantScopeWrite))
}
