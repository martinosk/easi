package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGrantScope_Valid(t *testing.T) {
	scope, err := NewGrantScope("write")
	require.NoError(t, err)
	assert.Equal(t, GrantScopeWrite, scope)
}

func TestNewGrantScope_Invalid(t *testing.T) {
	invalidScopes := []string{"", "read", "Write", "WRITE", "admin", "unknown"}

	for _, input := range invalidScopes {
		t.Run(input, func(t *testing.T) {
			_, err := NewGrantScope(input)
			assert.Error(t, err)
			assert.Equal(t, ErrInvalidGrantScope, err)
		})
	}
}

func TestGrantScope_String(t *testing.T) {
	assert.Equal(t, "write", GrantScopeWrite.String())
}

func TestGrantScope_Equals_SameValue(t *testing.T) {
	assert.True(t, GrantScopeWrite.Equals(GrantScopeWrite))
}

func TestGrantScope_Equals_DifferentValueObjectType(t *testing.T) {
	assert.False(t, GrantScopeWrite.Equals(GrantStatusActive))
}
