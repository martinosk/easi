package valueobjects

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInvitationID_GeneratesValidUUID(t *testing.T) {
	id := NewInvitationID()
	assert.NotEmpty(t, id.Value())

	_, err := uuid.Parse(id.Value())
	assert.NoError(t, err)
}

func TestNewInvitationID_GeneratesUniqueIDs(t *testing.T) {
	id1 := NewInvitationID()
	id2 := NewInvitationID()
	assert.NotEqual(t, id1.Value(), id2.Value())
}

func TestNewInvitationIDFromString_ValidUUID(t *testing.T) {
	validUUID := uuid.New().String()
	id, err := NewInvitationIDFromString(validUUID)
	require.NoError(t, err)
	assert.Equal(t, validUUID, id.Value())
}

func TestNewInvitationIDFromString_InvalidInputs(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "invalid format",
			input: "not-a-uuid",
		},
		{
			name:  "partial uuid",
			input: "123e4567-e89b-12d3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewInvitationIDFromString(tc.input)
			assert.ErrorIs(t, err, ErrInvalidInvitationID)
		})
	}
}

func TestInvitationID_String(t *testing.T) {
	id := NewInvitationID()
	assert.Equal(t, id.Value(), id.String())
}

func TestInvitationID_Equals(t *testing.T) {
	uuid1 := uuid.New().String()
	id1, _ := NewInvitationIDFromString(uuid1)
	id2, _ := NewInvitationIDFromString(uuid1)
	id3 := NewInvitationID()

	assert.True(t, id1.Equals(id2))
	assert.False(t, id1.Equals(id3))
}
