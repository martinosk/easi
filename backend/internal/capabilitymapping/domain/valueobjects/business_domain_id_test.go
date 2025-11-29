package valueobjects

import (
	"testing"

	"easi/backend/internal/shared/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewBusinessDomainID(t *testing.T) {
	id := NewBusinessDomainID()
	assert.NotEmpty(t, id.Value())
}

func TestNewBusinessDomainIDFromString_Valid(t *testing.T) {
	validID := "550e8400-e29b-41d4-a716-446655440000"
	id, err := NewBusinessDomainIDFromString(validID)
	assert.NoError(t, err)
	assert.Equal(t, validID, id.Value())
}

func TestNewBusinessDomainIDFromString_Empty(t *testing.T) {
	_, err := NewBusinessDomainIDFromString("")
	assert.Error(t, err)
	assert.Equal(t, domain.ErrEmptyValue, err)
}

func TestNewBusinessDomainIDFromString_InvalidGUID(t *testing.T) {
	_, err := NewBusinessDomainIDFromString("not-a-guid")
	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidValue, err)
}

func TestBusinessDomainID_String(t *testing.T) {
	id := NewBusinessDomainID()
	assert.Equal(t, id.Value(), id.String())
}

func TestBusinessDomainID_Equals(t *testing.T) {
	id1, _ := NewBusinessDomainIDFromString("550e8400-e29b-41d4-a716-446655440000")
	id2, _ := NewBusinessDomainIDFromString("550e8400-e29b-41d4-a716-446655440000")
	id3, _ := NewBusinessDomainIDFromString("660e8400-e29b-41d4-a716-446655440000")

	assert.True(t, id1.Equals(id2))
	assert.False(t, id1.Equals(id3))
}
