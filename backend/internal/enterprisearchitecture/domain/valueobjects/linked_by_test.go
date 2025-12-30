package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLinkedBy_ValidEmail(t *testing.T) {
	lb, err := NewLinkedBy("user@example.com")
	require.NoError(t, err)
	assert.Equal(t, "user@example.com", lb.Value())
	assert.False(t, lb.IsSystem())
}

func TestNewLinkedBy_SystemIdentifier(t *testing.T) {
	lb, err := NewLinkedBy("system")
	require.NoError(t, err)
	assert.Equal(t, "system", lb.Value())
	assert.True(t, lb.IsSystem())
}

func TestNewLinkedBy_SystemCaseInsensitive(t *testing.T) {
	lb, err := NewLinkedBy("SYSTEM")
	require.NoError(t, err)
	assert.Equal(t, "system", lb.Value())
	assert.True(t, lb.IsSystem())
}

func TestNewLinkedBy_NormalizesEmail(t *testing.T) {
	lb, err := NewLinkedBy("  User@Example.COM  ")
	require.NoError(t, err)
	assert.Equal(t, "user@example.com", lb.Value())
}

func TestNewLinkedBy_EmptyReturnsError(t *testing.T) {
	_, err := NewLinkedBy("")
	assert.ErrorIs(t, err, ErrLinkedByEmpty)
}

func TestNewLinkedBy_WhitespaceOnlyReturnsError(t *testing.T) {
	_, err := NewLinkedBy("   ")
	assert.ErrorIs(t, err, ErrLinkedByEmpty)
}

func TestNewLinkedBy_InvalidEmailReturnsError(t *testing.T) {
	testCases := []string{
		"notanemail",
		"missing@domain",
		"@nodomain.com",
		"spaces in@email.com",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			_, err := NewLinkedBy(tc)
			assert.ErrorIs(t, err, ErrLinkedByInvalid)
		})
	}
}

func TestNewLinkedBy_TooLongReturnsError(t *testing.T) {
	longEmail := make([]byte, 256)
	for i := range longEmail {
		longEmail[i] = 'a'
	}
	_, err := NewLinkedBy(string(longEmail))
	assert.ErrorIs(t, err, ErrLinkedByTooLong)
}

func TestLinkedBy_Equals(t *testing.T) {
	lb1, _ := NewLinkedBy("user@example.com")
	lb2, _ := NewLinkedBy("user@example.com")
	lb3, _ := NewLinkedBy("other@example.com")

	assert.True(t, lb1.Equals(lb2))
	assert.False(t, lb1.Equals(lb3))
}

func TestMustNewLinkedBy_PanicsOnInvalid(t *testing.T) {
	assert.Panics(t, func() {
		MustNewLinkedBy("")
	})
}

func TestMustNewLinkedBy_SucceedsOnValid(t *testing.T) {
	assert.NotPanics(t, func() {
		lb := MustNewLinkedBy("user@example.com")
		assert.Equal(t, "user@example.com", lb.Value())
	})
}
