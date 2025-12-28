package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBusinessDomainIDFromString_Empty(t *testing.T) {
	_, err := NewBusinessDomainIDFromString("")
	assert.Error(t, err)
	assert.Equal(t, domain.ErrEmptyValue, err)
}

func TestBusinessDomainID_Equals(t *testing.T) {
	id1, _ := NewBusinessDomainIDFromString("550e8400-e29b-41d4-a716-446655440000")
	id2, _ := NewBusinessDomainIDFromString("550e8400-e29b-41d4-a716-446655440000")
	id3, _ := NewBusinessDomainIDFromString("660e8400-e29b-41d4-a716-446655440000")

	assert.True(t, id1.Equals(id2))
	assert.False(t, id1.Equals(id3))
}
