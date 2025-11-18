package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLayoutDirection_TB(t *testing.T) {
	direction, err := NewLayoutDirection("TB")
	assert.NoError(t, err)
	assert.Equal(t, "TB", direction.Value())
}

func TestNewLayoutDirection_LR(t *testing.T) {
	direction, err := NewLayoutDirection("LR")
	assert.NoError(t, err)
	assert.Equal(t, "LR", direction.Value())
}

func TestNewLayoutDirection_BT(t *testing.T) {
	direction, err := NewLayoutDirection("BT")
	assert.NoError(t, err)
	assert.Equal(t, "BT", direction.Value())
}

func TestNewLayoutDirection_RL(t *testing.T) {
	direction, err := NewLayoutDirection("RL")
	assert.NoError(t, err)
	assert.Equal(t, "RL", direction.Value())
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

func TestNewLayoutDirection_LowerCase(t *testing.T) {
	_, err := NewLayoutDirection("tb")
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

func TestLayoutDirection_String(t *testing.T) {
	direction, _ := NewLayoutDirection("BT")
	assert.Equal(t, "BT", direction.String())
}
