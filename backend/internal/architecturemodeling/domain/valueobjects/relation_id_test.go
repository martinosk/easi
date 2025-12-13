package valueobjects

import (
	"testing"

	"easi/backend/internal/shared/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRelationID_GeneratesValidUUID(t *testing.T) {
	// Act
	id := NewRelationID()

	// Assert: Should generate a valid UUID
	assert.NotEmpty(t, id.Value())

	// Verify it's a valid UUID
	_, err := uuid.Parse(id.Value())
	assert.NoError(t, err)
}

func TestNewRelationID_GeneratesUniqueIDs(t *testing.T) {
	// Act: Generate multiple IDs
	id1 := NewRelationID()
	id2 := NewRelationID()
	id3 := NewRelationID()

	// Assert: All IDs should be unique
	assert.NotEqual(t, id1.Value(), id2.Value())
	assert.NotEqual(t, id1.Value(), id3.Value())
	assert.NotEqual(t, id2.Value(), id3.Value())
}

func TestNewRelationIDFromString_ValidUUID(t *testing.T) {
	// Arrange
	validUUID := uuid.New().String()

	// Act
	id, err := NewRelationIDFromString(validUUID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, validUUID, id.Value())
}

func TestNewRelationIDFromString_EmptyString(t *testing.T) {
	// Act
	_, err := NewRelationIDFromString("")

	// Assert: Should reject empty string
	assert.Error(t, err)
	assert.Equal(t, domain.ErrEmptyValue, err)
}

func TestNewRelationIDFromString_InvalidUUID(t *testing.T) {
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
		{"wrong format", "12345678-1234-1234-1234"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			_, err := NewRelationIDFromString(tc.invalidUUID)

			// Assert: Should reject invalid format
			assert.Error(t, err)
			assert.Equal(t, domain.ErrInvalidValue, err)
		})
	}
}

func TestRelationID_Equals(t *testing.T) {
	// Arrange
	uuidStr := uuid.New().String()
	id1, err := NewRelationIDFromString(uuidStr)
	require.NoError(t, err)

	id2, err := NewRelationIDFromString(uuidStr)
	require.NoError(t, err)

	differentUUID := uuid.New().String()
	id3, err := NewRelationIDFromString(differentUUID)
	require.NoError(t, err)

	// Act & Assert: Same UUIDs should be equal
	assert.True(t, id1.Equals(id2))
	assert.True(t, id2.Equals(id1))

	// Different UUIDs should not be equal
	assert.False(t, id1.Equals(id3))
	assert.False(t, id3.Equals(id1))
}

func TestRelationID_Equals_WithDifferentValueObjectType(t *testing.T) {
	// Arrange
	relationID := NewRelationID()

	// Create a different value object type (ComponentID) for comparison
	componentID := NewComponentID()

	// Act & Assert: Different value object types should not be equal
	assert.False(t, relationID.Equals(componentID))
}

func TestRelationID_String(t *testing.T) {
	// Arrange
	uuidStr := uuid.New().String()
	id, err := NewRelationIDFromString(uuidStr)
	require.NoError(t, err)

	// Act & Assert: String() should return the UUID value
	assert.Equal(t, uuidStr, id.String())
}

func TestRelationID_Value(t *testing.T) {
	// Arrange
	uuidStr := uuid.New().String()
	id, err := NewRelationIDFromString(uuidStr)
	require.NoError(t, err)

	// Act & Assert: Value() should return the underlying string
	assert.Equal(t, uuidStr, id.Value())
}

func TestRelationID_CanBeUsedAsMapKey(t *testing.T) {
	// Arrange
	id1 := NewRelationID()
	id2 := NewRelationID()

	// Act: Use RelationID as map key
	relationMap := make(map[RelationID]string)
	relationMap[id1] = "First relation"
	relationMap[id2] = "Second relation"

	// Assert: Map should store both relations
	assert.Len(t, relationMap, 2)
	assert.Equal(t, "First relation", relationMap[id1])
	assert.Equal(t, "Second relation", relationMap[id2])
}
