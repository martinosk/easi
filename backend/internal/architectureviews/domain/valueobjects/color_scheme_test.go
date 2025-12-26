package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewColorScheme_Valid(t *testing.T) {
	colorScheme, err := NewColorScheme("maturity")
	assert.NoError(t, err)
	assert.Equal(t, "maturity", colorScheme.Value())
}

func TestNewColorScheme_Invalid(t *testing.T) {
	_, err := NewColorScheme("invalid-scheme")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidColorScheme, err)
}

func TestNewColorScheme_EmptyString(t *testing.T) {
	_, err := NewColorScheme("")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidColorScheme, err)
}

func TestDefaultColorScheme(t *testing.T) {
	colorScheme := DefaultColorScheme()
	assert.Equal(t, "maturity", colorScheme.Value())
}

func TestColorScheme_Equals(t *testing.T) {
	colorScheme1, _ := NewColorScheme("maturity")
	colorScheme2, _ := NewColorScheme("maturity")
	colorScheme3, _ := NewColorScheme("classic")

	assert.True(t, colorScheme1.Equals(colorScheme2))
	assert.False(t, colorScheme1.Equals(colorScheme3))
}
