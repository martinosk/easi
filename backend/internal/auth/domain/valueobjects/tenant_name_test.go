package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTenantName_Valid(t *testing.T) {
	name, err := NewTenantName("Acme Corporation")
	assert.NoError(t, err)
	assert.Equal(t, "Acme Corporation", name.Value())
}

func TestNewTenantName_TrimSpace(t *testing.T) {
	name, err := NewTenantName("  Acme Corporation  ")
	assert.NoError(t, err)
	assert.Equal(t, "Acme Corporation", name.Value())
}

func TestNewTenantName_Empty(t *testing.T) {
	_, err := NewTenantName("")
	assert.Error(t, err)
	assert.Equal(t, ErrTenantNameEmpty, err)
}

func TestNewTenantName_OnlyWhitespace(t *testing.T) {
	_, err := NewTenantName("   ")
	assert.Error(t, err)
	assert.Equal(t, ErrTenantNameEmpty, err)
}

func TestNewTenantName_TooLong(t *testing.T) {
	longName := make([]byte, 256)
	for i := range longName {
		longName[i] = 'a'
	}
	_, err := NewTenantName(string(longName))
	assert.Error(t, err)
	assert.Equal(t, ErrTenantNameTooLong, err)
}

func TestNewTenantName_MaxLength(t *testing.T) {
	maxName := make([]byte, 255)
	for i := range maxName {
		maxName[i] = 'a'
	}
	name, err := NewTenantName(string(maxName))
	assert.NoError(t, err)
	assert.Equal(t, 255, len(name.Value()))
}

func TestTenantName_String(t *testing.T) {
	name, _ := NewTenantName("Acme Corporation")
	assert.Equal(t, "Acme Corporation", name.String())
}

func TestTenantName_Equals(t *testing.T) {
	name1, _ := NewTenantName("Acme Corporation")
	name2, _ := NewTenantName("Acme Corporation")
	name3, _ := NewTenantName("Other Company")

	assert.True(t, name1.Equals(name2))
	assert.False(t, name1.Equals(name3))
}
