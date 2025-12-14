package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvitationStatusFromString_ValidStatuses(t *testing.T) {
	testCases := []struct {
		input    string
		expected InvitationStatus
	}{
		{"pending", InvitationStatusPending},
		{"accepted", InvitationStatusAccepted},
		{"expired", InvitationStatusExpired},
		{"revoked", InvitationStatusRevoked},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			status, err := InvitationStatusFromString(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, status)
		})
	}
}

func TestInvitationStatusFromString_InvalidStatus(t *testing.T) {
	testCases := []string{
		"active",
		"cancelled",
		"PENDING",
		"",
	}

	for _, input := range testCases {
		t.Run(input, func(t *testing.T) {
			_, err := InvitationStatusFromString(input)
			assert.ErrorIs(t, err, ErrInvalidInvitationStatus)
		})
	}
}

func TestInvitationStatus_StringRepresentation(t *testing.T) {
	assert.Equal(t, "pending", InvitationStatusPending.String())
	assert.Equal(t, "accepted", InvitationStatusAccepted.String())
	assert.Equal(t, "expired", InvitationStatusExpired.String())
	assert.Equal(t, "revoked", InvitationStatusRevoked.String())
}

func TestInvitationStatus_StateChecks(t *testing.T) {
	t.Run("pending checks", func(t *testing.T) {
		assert.True(t, InvitationStatusPending.IsPending())
		assert.False(t, InvitationStatusPending.IsAccepted())
		assert.False(t, InvitationStatusPending.IsExpired())
		assert.False(t, InvitationStatusPending.IsRevoked())
	})

	t.Run("accepted checks", func(t *testing.T) {
		assert.False(t, InvitationStatusAccepted.IsPending())
		assert.True(t, InvitationStatusAccepted.IsAccepted())
		assert.False(t, InvitationStatusAccepted.IsExpired())
		assert.False(t, InvitationStatusAccepted.IsRevoked())
	})

	t.Run("expired checks", func(t *testing.T) {
		assert.False(t, InvitationStatusExpired.IsPending())
		assert.False(t, InvitationStatusExpired.IsAccepted())
		assert.True(t, InvitationStatusExpired.IsExpired())
		assert.False(t, InvitationStatusExpired.IsRevoked())
	})

	t.Run("revoked checks", func(t *testing.T) {
		assert.False(t, InvitationStatusRevoked.IsPending())
		assert.False(t, InvitationStatusRevoked.IsAccepted())
		assert.False(t, InvitationStatusRevoked.IsExpired())
		assert.True(t, InvitationStatusRevoked.IsRevoked())
	})
}

func TestInvitationStatus_CanTransitionTo(t *testing.T) {
	t.Run("pending can transition to accepted, expired, revoked", func(t *testing.T) {
		assert.True(t, InvitationStatusPending.CanTransitionTo(InvitationStatusAccepted))
		assert.True(t, InvitationStatusPending.CanTransitionTo(InvitationStatusExpired))
		assert.True(t, InvitationStatusPending.CanTransitionTo(InvitationStatusRevoked))
		assert.False(t, InvitationStatusPending.CanTransitionTo(InvitationStatusPending))
	})

	t.Run("accepted cannot transition", func(t *testing.T) {
		assert.False(t, InvitationStatusAccepted.CanTransitionTo(InvitationStatusPending))
		assert.False(t, InvitationStatusAccepted.CanTransitionTo(InvitationStatusExpired))
		assert.False(t, InvitationStatusAccepted.CanTransitionTo(InvitationStatusRevoked))
	})

	t.Run("expired cannot transition", func(t *testing.T) {
		assert.False(t, InvitationStatusExpired.CanTransitionTo(InvitationStatusPending))
		assert.False(t, InvitationStatusExpired.CanTransitionTo(InvitationStatusAccepted))
		assert.False(t, InvitationStatusExpired.CanTransitionTo(InvitationStatusRevoked))
	})

	t.Run("revoked cannot transition", func(t *testing.T) {
		assert.False(t, InvitationStatusRevoked.CanTransitionTo(InvitationStatusPending))
		assert.False(t, InvitationStatusRevoked.CanTransitionTo(InvitationStatusAccepted))
		assert.False(t, InvitationStatusRevoked.CanTransitionTo(InvitationStatusExpired))
	})
}

func TestInvitationStatus_Equals(t *testing.T) {
	status1 := InvitationStatusPending
	status2, _ := InvitationStatusFromString("pending")
	status3 := InvitationStatusAccepted

	assert.True(t, status1.Equals(status2))
	assert.False(t, status1.Equals(status3))
}
