package valueobjects

import (
	"testing"

	"easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
)

func TestNewRelationIDFromString_EmptyString(t *testing.T) {
	_, err := NewRelationIDFromString("")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrEmptyValue, err)
}

func TestRelationID_Equals(t *testing.T) {
	id1, _ := NewRelationIDFromString("550e8400-e29b-41d4-a716-446655440000")
	id2, _ := NewRelationIDFromString("550e8400-e29b-41d4-a716-446655440000")
	id3, _ := NewRelationIDFromString("660e8400-e29b-41d4-a716-446655440000")

	assert.True(t, id1.Equals(id2))
	assert.False(t, id1.Equals(id3))
}
