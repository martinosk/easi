package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewViewName_EmptyString(t *testing.T) {
	_, err := NewViewName("")
	assert.Error(t, err)
	assert.Equal(t, ErrViewNameEmpty, err)
}

func TestNewViewName_WhitespaceOnly(t *testing.T) {
	_, err := NewViewName("   ")
	assert.Error(t, err)
	assert.Equal(t, ErrViewNameEmpty, err)
}

func TestNewViewName_ExceedsMaxLength(t *testing.T) {
	tooLongName := "This is a view name that is one hundred and one characters long and should fail the validation tests!"
	_, err := NewViewName(tooLongName)
	assert.Error(t, err)
	assert.Equal(t, ErrViewNameTooLong, err)
}

func TestViewName_Equals(t *testing.T) {
	name1, _ := NewViewName("Dashboard")
	name2, _ := NewViewName("Dashboard")
	name3, _ := NewViewName("Reports")

	assert.True(t, name1.Equals(name2))
	assert.False(t, name1.Equals(name3))
}
