package valueobjects

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUserIdentifier_ValidValue(t *testing.T) {
	identifier, err := NewUserIdentifier("user@example.com")
	assert.NoError(t, err)
	assert.Equal(t, "user@example.com", identifier.Value())
}

func TestNewUserIdentifier_Empty(t *testing.T) {
	_, err := NewUserIdentifier("")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUserIdentifierEmpty)
}

func TestNewUserIdentifier_WhitespaceOnly(t *testing.T) {
	_, err := NewUserIdentifier("   ")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUserIdentifierEmpty)
}

func TestNewUserIdentifier_TrimsWhitespace(t *testing.T) {
	identifier, err := NewUserIdentifier("  john.doe@example.com  ")
	assert.NoError(t, err)
	assert.Equal(t, "john.doe@example.com", identifier.Value())
}

func TestNewUserIdentifier_MaxLength(t *testing.T) {
	text := strings.Repeat("a", 255)
	identifier, err := NewUserIdentifier(text)
	assert.NoError(t, err)
	assert.Equal(t, 255, len(identifier.Value()))
}

func TestNewUserIdentifier_ExceedsMaxLength(t *testing.T) {
	text := strings.Repeat("a", 256)
	_, err := NewUserIdentifier(text)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUserIdentifierTooLong)
}

func TestUserIdentifier_Equals(t *testing.T) {
	u1, _ := NewUserIdentifier("user@example.com")
	u2, _ := NewUserIdentifier("user@example.com")
	u3, _ := NewUserIdentifier("other@example.com")

	assert.True(t, u1.Equals(u2))
	assert.False(t, u1.Equals(u3))
}

func TestUserIdentifier_Equals_DifferentType(t *testing.T) {
	identifier, _ := NewUserIdentifier("user@example.com")
	score, _ := NewFitScore(3)

	assert.False(t, identifier.Equals(score))
}

func TestUserIdentifier_String(t *testing.T) {
	identifier, _ := NewUserIdentifier("user@example.com")
	assert.Equal(t, "user@example.com", identifier.String())
}
