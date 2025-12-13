package valueobjects

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNonce_GeneratesValidNonce(t *testing.T) {
	nonce := NewNonce()

	value := nonce.Value()
	assert.NotEmpty(t, value, "nonce should not be empty")

	urlSafe := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	assert.True(t, urlSafe.MatchString(value), "nonce should be URL-safe")
}

func TestNewNonce_GeneratesUniqueValues(t *testing.T) {
	nonce1 := NewNonce()
	nonce2 := NewNonce()

	assert.NotEqual(t, nonce1.Value(), nonce2.Value(), "each nonce should be unique")
}

func TestNonceFromValue_ValidNonce(t *testing.T) {
	original := NewNonce()

	restored, err := NonceFromValue(original.Value())
	require.NoError(t, err)

	assert.Equal(t, original.Value(), restored.Value())
}

func TestNonceFromValue_EmptyNonce(t *testing.T) {
	_, err := NonceFromValue("")
	assert.Error(t, err)
}

func TestNonce_Equals(t *testing.T) {
	nonce1 := NewNonce()

	nonce2, err := NonceFromValue(nonce1.Value())
	require.NoError(t, err)

	nonce3 := NewNonce()

	assert.True(t, nonce1.Equals(nonce2), "same value nonces should be equal")
	assert.False(t, nonce1.Equals(nonce3), "different nonces should not be equal")
}
