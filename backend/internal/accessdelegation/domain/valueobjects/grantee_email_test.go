package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGranteeEmail_EmptyEmail_ReturnsError(t *testing.T) {
	_, err := NewGranteeEmail("")
	assert.Equal(t, ErrGranteeEmailEmpty, err)
}

func TestNewGranteeEmail_WhitespaceOnly_ReturnsError(t *testing.T) {
	_, err := NewGranteeEmail("   ")
	assert.Equal(t, ErrGranteeEmailEmpty, err)
}

func TestNewGranteeEmail_ValidEmail_Succeeds(t *testing.T) {
	email, err := NewGranteeEmail("user@example.com")
	require.NoError(t, err)
	assert.Equal(t, "user@example.com", email.Value())
}

func TestNewGranteeEmail_TrimsAndLowercases(t *testing.T) {
	email, err := NewGranteeEmail("  User@Example.COM  ")
	require.NoError(t, err)
	assert.Equal(t, "user@example.com", email.Value())
}

func TestGranteeEmail_Equals_CaseInsensitive(t *testing.T) {
	e1, _ := NewGranteeEmail("User@Example.com")
	e2, _ := NewGranteeEmail("user@example.COM")
	assert.True(t, e1.Equals(e2))
}

func TestGranteeEmail_Equals_DifferentEmails_ReturnsFalse(t *testing.T) {
	e1, _ := NewGranteeEmail("user1@example.com")
	e2, _ := NewGranteeEmail("user2@example.com")
	assert.False(t, e1.Equals(e2))
}

func TestGranteeEmail_Equals_DifferentValueObjectType_ReturnsFalse(t *testing.T) {
	e, _ := NewGranteeEmail("user@example.com")
	assert.False(t, e.Equals(GrantScopeWrite))
}

func TestNewGranteeEmail_InvalidFormat_ReturnsError(t *testing.T) {
	_, err := NewGranteeEmail("not-an-email")
	assert.Equal(t, ErrGranteeEmailInvalid, err)
}
