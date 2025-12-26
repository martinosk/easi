package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserIDFromString_EmptyString(t *testing.T) {
	_, err := UserIDFromString("")

	assert.ErrorIs(t, err, ErrInvalidUserID)
}

func TestUserID_Equals(t *testing.T) {
	id := "550e8400-e29b-41d4-a716-446655440000"
	userID1, _ := UserIDFromString(id)
	userID2, _ := UserIDFromString(id)
	userID3, _ := UserIDFromString("660e8400-e29b-41d4-a716-446655440000")

	assert.True(t, userID1.Equals(userID2))
	assert.False(t, userID1.Equals(userID3))
}
