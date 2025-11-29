package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBusinessDomainID(t *testing.T) {
	id := NewBusinessDomainID()
	assert.NotEmpty(t, id.Value())
	assert.Contains(t, id.Value(), "bd-")
}

func TestNewBusinessDomainIDFromString_Valid(t *testing.T) {
	validID := "bd-550e8400-e29b-41d4-a716-446655440000"
	id, err := NewBusinessDomainIDFromString(validID)
	assert.NoError(t, err)
	assert.Equal(t, validID, id.Value())
}

func TestNewBusinessDomainIDFromString_Empty(t *testing.T) {
	_, err := NewBusinessDomainIDFromString("")
	assert.Error(t, err)
}

func TestNewBusinessDomainIDFromString_MissingPrefix(t *testing.T) {
	_, err := NewBusinessDomainIDFromString("550e8400-e29b-41d4-a716-446655440000")
	assert.Error(t, err)
	assert.Equal(t, ErrBusinessDomainIDMissingPrefix, err)
}

func TestNewBusinessDomainIDFromString_InvalidGUID(t *testing.T) {
	_, err := NewBusinessDomainIDFromString("bd-not-a-guid")
	assert.Error(t, err)
}

func TestNewBusinessDomainIDFromString_WrongPrefix(t *testing.T) {
	_, err := NewBusinessDomainIDFromString("cap-550e8400-e29b-41d4-a716-446655440000")
	assert.Error(t, err)
	assert.Equal(t, ErrBusinessDomainIDMissingPrefix, err)
}

func TestBusinessDomainID_String(t *testing.T) {
	id := NewBusinessDomainID()
	assert.Equal(t, id.Value(), id.String())
}

func TestBusinessDomainID_Equals(t *testing.T) {
	id1, _ := NewBusinessDomainIDFromString("bd-550e8400-e29b-41d4-a716-446655440000")
	id2, _ := NewBusinessDomainIDFromString("bd-550e8400-e29b-41d4-a716-446655440000")
	id3, _ := NewBusinessDomainIDFromString("bd-660e8400-e29b-41d4-a716-446655440000")

	assert.True(t, id1.Equals(id2))
	assert.False(t, id1.Equals(id3))
}
