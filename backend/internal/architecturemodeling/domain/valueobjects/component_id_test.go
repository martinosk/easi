package valueobjects

import (
	"testing"

	"easi/backend/internal/shared/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewComponentID_GeneratesValidUUID(t *testing.T) {
	// Act
	id := NewComponentID()

	// Assert: Should generate a valid UUID
	assert.NotEmpty(t, id.Value())

	// Verify it's a valid UUID
	_, err := uuid.Parse(id.Value())
	assert.NoError(t, err)
}

func TestNewComponentID_GeneratesUniqueIDs(t *testing.T) {
	// Act: Generate multiple IDs
	id1 := NewComponentID()
	id2 := NewComponentID()
	id3 := NewComponentID()

	// Assert: All IDs should be unique
	assert.NotEqual(t, id1.Value(), id2.Value())
	assert.NotEqual(t, id1.Value(), id3.Value())
	assert.NotEqual(t, id2.Value(), id3.Value())
}

func TestNewComponentIDFromString_ValidUUID(t *testing.T) {
	// Arrange
	validUUID := uuid.New().String()

	// Act
	id, err := NewComponentIDFromString(validUUID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, validUUID, id.Value())
}

func TestNewComponentIDFromString_EmptyString(t *testing.T) {
	// Act
	_, err := NewComponentIDFromString("")

	// Assert: Should reject empty string
	assert.Error(t, err)
	assert.Equal(t, domain.ErrEmptyValue, err)
}

func TestNewComponentIDFromString_InvalidUUID(t *testing.T) {
	// Arrange: Test various invalid UUID formats
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
			// Act
			_, err := NewComponentIDFromString(tc.invalidUUID)

			// Assert: Should reject invalid format
			assert.Error(t, err)
			assert.Equal(t, domain.ErrInvalidValue, err)
		})
	}
}

func TestComponentID_Equals(t *testing.T) {
	// Arrange
	uuidStr := uuid.New().String()
	id1, err := NewComponentIDFromString(uuidStr)
	require.NoError(t, err)

	id2, err := NewComponentIDFromString(uuidStr)
	require.NoError(t, err)

	differentUUID := uuid.New().String()
	id3, err := NewComponentIDFromString(differentUUID)
	require.NoError(t, err)

	// Act & Assert: Same UUIDs should be equal
	assert.True(t, id1.Equals(id2))
	assert.True(t, id2.Equals(id1))

	// Different UUIDs should not be equal
	assert.False(t, id1.Equals(id3))
	assert.False(t, id3.Equals(id1))
}

func TestComponentID_Equals_WithDifferentValueObjectType(t *testing.T) {
	// Arrange
	componentID := NewComponentID()

	// Create a different value object type (RelationID) for comparison
	relationID := NewRelationID()

	// Act & Assert: Different value object types should not be equal
	assert.False(t, componentID.Equals(relationID))
}

func TestComponentID_String(t *testing.T) {
	// Arrange
	uuidStr := uuid.New().String()
	id, err := NewComponentIDFromString(uuidStr)
	require.NoError(t, err)

	// Act & Assert: String() should return the UUID value
	assert.Equal(t, uuidStr, id.String())
}

func TestComponentID_Value(t *testing.T) {
	// Arrange
	uuidStr := uuid.New().String()
	id, err := NewComponentIDFromString(uuidStr)
	require.NoError(t, err)

	// Act & Assert: Value() should return the underlying string
	assert.Equal(t, uuidStr, id.Value())
}
