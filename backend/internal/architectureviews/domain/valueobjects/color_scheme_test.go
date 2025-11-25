package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewColorScheme_Maturity(t *testing.T) {
	colorScheme, err := NewColorScheme("maturity")
	assert.NoError(t, err)
	assert.Equal(t, "maturity", colorScheme.Value())
}

func TestNewColorScheme_Archimate(t *testing.T) {
	colorScheme, err := NewColorScheme("archimate")
	assert.NoError(t, err)
	assert.Equal(t, "archimate", colorScheme.Value())
}

func TestNewColorScheme_ArchimateClassic(t *testing.T) {
	colorScheme, err := NewColorScheme("archimate-classic")
	assert.NoError(t, err)
	assert.Equal(t, "archimate-classic", colorScheme.Value())
}

func TestNewColorScheme_Custom(t *testing.T) {
	colorScheme, err := NewColorScheme("custom")
	assert.NoError(t, err)
	assert.Equal(t, "custom", colorScheme.Value())
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

func TestNewColorScheme_CaseSensitive(t *testing.T) {
	_, err := NewColorScheme("MATURITY")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidColorScheme, err)
}

func TestNewColorScheme_PartialMatch(t *testing.T) {
	_, err := NewColorScheme("arch")
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
	colorScheme3, _ := NewColorScheme("archimate")

	assert.True(t, colorScheme1.Equals(colorScheme2))
	assert.False(t, colorScheme1.Equals(colorScheme3))
}

func TestColorScheme_String(t *testing.T) {
	colorScheme, _ := NewColorScheme("archimate-classic")
	assert.Equal(t, "archimate-classic", colorScheme.String())
}
