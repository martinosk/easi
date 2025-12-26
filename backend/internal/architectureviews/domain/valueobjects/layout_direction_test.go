package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLayoutDirection_Valid(t *testing.T) {
	direction, err := NewLayoutDirection("TB")
	assert.NoError(t, err)
	assert.Equal(t, "TB", direction.Value())
}

func TestNewLayoutDirection_Invalid(t *testing.T) {
	_, err := NewLayoutDirection("UP")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidLayoutDirection, err)
}

func TestNewLayoutDirection_EmptyString(t *testing.T) {
	_, err := NewLayoutDirection("")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidLayoutDirection, err)
}

func TestDefaultLayoutDirection(t *testing.T) {
	direction := DefaultLayoutDirection()
	assert.Equal(t, "TB", direction.Value())
}

func TestLayoutDirection_Equals(t *testing.T) {
	direction1, _ := NewLayoutDirection("TB")
	direction2, _ := NewLayoutDirection("TB")
	direction3, _ := NewLayoutDirection("LR")

	assert.True(t, direction1.Equals(direction2))
	assert.False(t, direction1.Equals(direction3))
}
