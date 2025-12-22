package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUserID_GeneratesValidUUID(t *testing.T) {
	userID := NewUserID()

	assert.NotEmpty(t, userID.Value())
	assert.NotEmpty(t, userID.String())
	assert.Equal(t, userID.Value(), userID.String())
}

func TestNewUserID_GeneratesUniqueIDs(t *testing.T) {
	id1 := NewUserID()
	id2 := NewUserID()

	assert.NotEqual(t, id1.Value(), id2.Value())
}

func TestUserIDFromString_ValidUUID(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	userID, err := UserIDFromString(validUUID)

	require.NoError(t, err)
	assert.Equal(t, validUUID, userID.Value())
}

func TestUserIDFromString_EmptyString(t *testing.T) {
	_, err := UserIDFromString("")

	assert.ErrorIs(t, err, ErrInvalidUserID)
}

func TestUserIDFromString_InvalidUUID(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{"not a UUID", "not-a-uuid"},
		{"partial UUID", "550e8400-e29b-41d4"},
		{"invalid characters", "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"},
		{"spaces", "550e8400 e29b 41d4 a716 446655440000"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := UserIDFromString(tc.input)
			assert.ErrorIs(t, err, ErrInvalidUserID)
		})
	}
}

func TestUserID_Equals(t *testing.T) {
	id := "550e8400-e29b-41d4-a716-446655440000"
	userID1, _ := UserIDFromString(id)
	userID2, _ := UserIDFromString(id)
	userID3, _ := UserIDFromString("660e8400-e29b-41d4-a716-446655440000")

	assert.True(t, userID1.Equals(userID2))
	assert.False(t, userID1.Equals(userID3))
}

func TestUserID_EqualsWithDifferentType(t *testing.T) {
	userID := NewUserID()
	email, _ := NewEmail("test@example.com")

	assert.False(t, userID.Equals(email))
}
