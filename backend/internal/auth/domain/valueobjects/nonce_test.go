package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNonceFromValue_EmptyNonce(t *testing.T) {
	_, err := NonceFromValue("")
	assert.Error(t, err)
}

func TestNonce_Equals(t *testing.T) {
	nonce1 := NewNonce()

	nonce2, _ := NonceFromValue(nonce1.Value())
	nonce3 := NewNonce()

	assert.True(t, nonce1.Equals(nonce2))
	assert.False(t, nonce1.Equals(nonce3))
}
