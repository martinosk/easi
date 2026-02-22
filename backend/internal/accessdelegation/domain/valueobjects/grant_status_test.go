package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGrantStatus_ValidStatuses(t *testing.T) {
	tests := []struct {
		input    string
		expected GrantStatus
	}{
		{"active", GrantStatusActive},
		{"revoked", GrantStatusRevoked},
		{"expired", GrantStatusExpired},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			status, err := NewGrantStatus(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, status)
		})
	}
}

func TestNewGrantStatus_Invalid(t *testing.T) {
	invalidStatuses := []string{"", "Active", "REVOKED", "pending", "unknown"}

	for _, input := range invalidStatuses {
		t.Run(input, func(t *testing.T) {
			_, err := NewGrantStatus(input)
			assert.Error(t, err)
			assert.Equal(t, ErrInvalidGrantStatus, err)
		})
	}
}

func TestGrantStatus_String(t *testing.T) {
	assert.Equal(t, "active", GrantStatusActive.String())
	assert.Equal(t, "revoked", GrantStatusRevoked.String())
	assert.Equal(t, "expired", GrantStatusExpired.String())
}

func TestGrantStatus_IsActive(t *testing.T) {
	assert.True(t, GrantStatusActive.IsActive())
	assert.False(t, GrantStatusRevoked.IsActive())
	assert.False(t, GrantStatusExpired.IsActive())
}

func TestGrantStatus_IsRevoked(t *testing.T) {
	assert.False(t, GrantStatusActive.IsRevoked())
	assert.True(t, GrantStatusRevoked.IsRevoked())
	assert.False(t, GrantStatusExpired.IsRevoked())
}

func TestGrantStatus_IsExpired(t *testing.T) {
	assert.False(t, GrantStatusActive.IsExpired())
	assert.False(t, GrantStatusRevoked.IsExpired())
	assert.True(t, GrantStatusExpired.IsExpired())
}

func TestGrantStatus_Equals_SameValue(t *testing.T) {
	activeA, activeB := GrantStatusActive, GrantStatusActive
	revokedA, revokedB := GrantStatusRevoked, GrantStatusRevoked
	expiredA, expiredB := GrantStatusExpired, GrantStatusExpired
	assert.True(t, activeA.Equals(activeB))
	assert.True(t, revokedA.Equals(revokedB))
	assert.True(t, expiredA.Equals(expiredB))
}

func TestGrantStatus_Equals_DifferentValue(t *testing.T) {
	assert.False(t, GrantStatusActive.Equals(GrantStatusRevoked))
	assert.False(t, GrantStatusActive.Equals(GrantStatusExpired))
	assert.False(t, GrantStatusRevoked.Equals(GrantStatusExpired))
}

func TestGrantStatus_Equals_DifferentValueObjectType(t *testing.T) {
	ref, _ := NewArtifactRef(ArtifactTypeCapability, "cap-123")
	assert.False(t, GrantStatusActive.Equals(ref))
}
