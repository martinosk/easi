package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserStatusFromString_Valid(t *testing.T) {
	status, err := UserStatusFromString("active")
	require.NoError(t, err)
	assert.Equal(t, UserStatusActive, status)
}

func TestUserStatusFromString_InvalidStatus(t *testing.T) {
	_, err := UserStatusFromString("banned")
	assert.ErrorIs(t, err, ErrInvalidUserStatus)
}

func TestUserStatusFromString_EmptyStatus(t *testing.T) {
	_, err := UserStatusFromString("")
	assert.ErrorIs(t, err, ErrInvalidUserStatus)
}

func TestUserStatus_IsActive(t *testing.T) {
	assert.True(t, UserStatusActive.IsActive())
	assert.False(t, UserStatusDisabled.IsActive())
}
