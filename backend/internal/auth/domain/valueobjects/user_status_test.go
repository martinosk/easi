package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserStatusFromString_ValidStatuses(t *testing.T) {
	testCases := []struct {
		input    string
		expected UserStatus
	}{
		{"active", UserStatusActive},
		{"disabled", UserStatusDisabled},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			status, err := UserStatusFromString(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, status)
		})
	}
}

func TestUserStatusFromString_InvalidStatus(t *testing.T) {
	_, err := UserStatusFromString("banned")
	assert.ErrorIs(t, err, ErrInvalidUserStatus)
}

func TestUserStatusFromString_EmptyStatus(t *testing.T) {
	_, err := UserStatusFromString("")
	assert.ErrorIs(t, err, ErrInvalidUserStatus)
}

func TestUserStatus_String(t *testing.T) {
	assert.Equal(t, "active", UserStatusActive.String())
	assert.Equal(t, "disabled", UserStatusDisabled.String())
}

func TestUserStatus_IsActive(t *testing.T) {
	assert.True(t, UserStatusActive.IsActive())
	assert.False(t, UserStatusDisabled.IsActive())
}

func TestUserStatus_Equals(t *testing.T) {
	status1 := UserStatusActive
	status2, _ := UserStatusFromString("active")
	status3 := UserStatusDisabled

	assert.True(t, status1.Equals(status2), "same statuses should be equal")
	assert.False(t, status1.Equals(status3), "different statuses should not be equal")
}
