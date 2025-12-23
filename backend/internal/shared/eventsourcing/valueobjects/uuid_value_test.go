package valueobjects

import (
	"testing"

	"easi/backend/internal/shared/eventsourcing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUUIDValue_GeneratesValidUUID(t *testing.T) {
	id := NewUUIDValue()

	assert.NotEmpty(t, id.Value())

	_, err := uuid.Parse(id.Value())
	assert.NoError(t, err)
}

func TestNewUUIDValue_GeneratesUniqueIDs(t *testing.T) {
	id1 := NewUUIDValue()
	id2 := NewUUIDValue()
	id3 := NewUUIDValue()

	assert.NotEqual(t, id1.Value(), id2.Value())
	assert.NotEqual(t, id1.Value(), id3.Value())
	assert.NotEqual(t, id2.Value(), id3.Value())
}

func TestNewUUIDValueFromString_ValidUUID(t *testing.T) {
	validUUID := uuid.New().String()

	id, err := NewUUIDValueFromString(validUUID)

	require.NoError(t, err)
	assert.Equal(t, validUUID, id.Value())
}

func TestNewUUIDValueFromString_EmptyString(t *testing.T) {
	_, err := NewUUIDValueFromString("")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrEmptyValue, err)
}

func TestNewUUIDValueFromString_InvalidUUID(t *testing.T) {
	testCases := []struct {
		name        string
		invalidUUID string
	}{
		{"random string", "not-a-uuid"},
		{"partial UUID", "123e4567-e89b"},
		{"numeric only", "12345678"},
		{"special characters", "!@#$%^&*()"},
		{"whitespace", "   "},
		{"almost valid", "123e4567-e89b-12d3-a456"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewUUIDValueFromString(tc.invalidUUID)

			assert.Error(t, err)
			assert.Equal(t, domain.ErrInvalidValue, err)
		})
	}
}

func TestUUIDValue_EqualsValue(t *testing.T) {
	uuidStr := uuid.New().String()
	id1, err := NewUUIDValueFromString(uuidStr)
	require.NoError(t, err)

	id2, err := NewUUIDValueFromString(uuidStr)
	require.NoError(t, err)

	differentUUID := uuid.New().String()
	id3, err := NewUUIDValueFromString(differentUUID)
	require.NoError(t, err)

	assert.True(t, id1.EqualsValue(id2))
	assert.True(t, id2.EqualsValue(id1))

	assert.False(t, id1.EqualsValue(id3))
	assert.False(t, id3.EqualsValue(id1))
}

func TestUUIDValue_String(t *testing.T) {
	uuidStr := uuid.New().String()
	id, err := NewUUIDValueFromString(uuidStr)
	require.NoError(t, err)

	assert.Equal(t, uuidStr, id.String())
}

func TestUUIDValue_Value(t *testing.T) {
	uuidStr := uuid.New().String()
	id, err := NewUUIDValueFromString(uuidStr)
	require.NoError(t, err)

	assert.Equal(t, uuidStr, id.Value())
}
