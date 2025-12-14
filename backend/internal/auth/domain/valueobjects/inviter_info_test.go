package valueobjects

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInviterInfo_ValidInputs(t *testing.T) {
	userID := uuid.New()
	email, _ := NewEmail("admin@example.com")

	info, err := NewInviterInfo(userID, email)
	require.NoError(t, err)
	assert.Equal(t, userID, info.UserID())
	assert.Equal(t, email, info.Email())
}

func TestNewInviterInfo_NilUserID(t *testing.T) {
	email, _ := NewEmail("admin@example.com")

	_, err := NewInviterInfo(uuid.Nil, email)
	assert.ErrorIs(t, err, ErrInvalidInviterID)
}

func TestNewInviterInfo_EmptyEmail(t *testing.T) {
	userID := uuid.New()
	emptyEmail := Email{}

	_, err := NewInviterInfo(userID, emptyEmail)
	assert.ErrorIs(t, err, ErrInvalidInviterEmail)
}

func TestInviterInfo_Equals(t *testing.T) {
	userID := uuid.New()
	email, _ := NewEmail("admin@example.com")

	info1, _ := NewInviterInfo(userID, email)
	info2, _ := NewInviterInfo(userID, email)

	otherUserID := uuid.New()
	info3, _ := NewInviterInfo(otherUserID, email)

	assert.True(t, info1.Equals(info2))
	assert.False(t, info1.Equals(info3))
}
