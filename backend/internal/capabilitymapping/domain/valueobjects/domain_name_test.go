package valueobjects

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestNewDomainName_ExceedsMaxLength(t *testing.T) {
	tooLongName := strings.Repeat("a", 101)
	_, err := NewDomainName(tooLongName)
	assert.Error(t, err)
	assert.Equal(t, ErrDomainNameTooLong, err)
}

func TestDomainName_Equals(t *testing.T) {
	name1, _ := NewDomainName("Finance")
	name2, _ := NewDomainName("Finance")
	name3, _ := NewDomainName("Operations")

	assert.True(t, name1.Equals(name2))
	assert.False(t, name1.Equals(name3))
}
