package valueobjects

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDomainName_Valid(t *testing.T) {
	name, err := NewDomainName("Finance")
	assert.NoError(t, err)
	assert.Equal(t, "Finance", name.Value())
}

func TestNewDomainName_TrimSpace(t *testing.T) {
	name, err := NewDomainName("  Customer Experience  ")
	assert.NoError(t, err)
	assert.Equal(t, "Customer Experience", name.Value())
}

func TestNewDomainName_Empty(t *testing.T) {
	_, err := NewDomainName("")
	assert.Error(t, err)
	assert.Equal(t, ErrDomainNameEmpty, err)
}

func TestNewDomainName_OnlyWhitespace(t *testing.T) {
	_, err := NewDomainName("   ")
	assert.Error(t, err)
	assert.Equal(t, ErrDomainNameEmpty, err)
}

func TestNewDomainName_MaxLength(t *testing.T) {
	validName := strings.Repeat("a", 100)
	name, err := NewDomainName(validName)
	assert.NoError(t, err)
	assert.Equal(t, validName, name.Value())
}

func TestNewDomainName_ExceedsMaxLength(t *testing.T) {
	tooLongName := strings.Repeat("a", 101)
	_, err := NewDomainName(tooLongName)
	assert.Error(t, err)
	assert.Equal(t, ErrDomainNameTooLong, err)
}

func TestDomainName_String(t *testing.T) {
	name, _ := NewDomainName("Operations")
	assert.Equal(t, "Operations", name.String())
}

func TestDomainName_Equals(t *testing.T) {
	name1, _ := NewDomainName("Finance")
	name2, _ := NewDomainName("Finance")
	name3, _ := NewDomainName("Operations")

	assert.True(t, name1.Equals(name2))
	assert.False(t, name1.Equals(name3))
}
