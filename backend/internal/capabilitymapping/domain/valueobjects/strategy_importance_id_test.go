package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStrategyImportanceID_GeneratesValidID(t *testing.T) {
	id := NewStrategyImportanceID()
	assert.NotEmpty(t, id.Value())
}

func TestNewStrategyImportanceIDFromString_ValidUUID(t *testing.T) {
	id, err := NewStrategyImportanceIDFromString("550e8400-e29b-41d4-a716-446655440000")
	assert.NoError(t, err)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", id.Value())
}

func TestNewStrategyImportanceIDFromString_InvalidUUID(t *testing.T) {
	_, err := NewStrategyImportanceIDFromString("invalid")
	assert.Error(t, err)
}

func TestNewStrategyImportanceIDFromString_Empty(t *testing.T) {
	_, err := NewStrategyImportanceIDFromString("")
	assert.Error(t, err)
}

func TestStrategyImportanceID_Equals(t *testing.T) {
	id1, _ := NewStrategyImportanceIDFromString("550e8400-e29b-41d4-a716-446655440000")
	id2, _ := NewStrategyImportanceIDFromString("550e8400-e29b-41d4-a716-446655440000")
	id3, _ := NewStrategyImportanceIDFromString("660e8400-e29b-41d4-a716-446655440000")

	assert.True(t, id1.Equals(id2))
	assert.False(t, id1.Equals(id3))
}
