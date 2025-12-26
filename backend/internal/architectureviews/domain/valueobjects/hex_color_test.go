package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHexColor_Valid(t *testing.T) {
	hexColor, err := NewHexColor("#FFFFFF")
	assert.NoError(t, err)
	assert.Equal(t, "#FFFFFF", hexColor.Value())
}

func TestNewHexColor_MissingHashPrefix(t *testing.T) {
	_, err := NewHexColor("FFFFFF")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidHexColor, err)
}

func TestNewHexColor_TooShort(t *testing.T) {
	_, err := NewHexColor("#FFF")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidHexColor, err)
}

func TestNewHexColor_InvalidCharacters(t *testing.T) {
	_, err := NewHexColor("#GGGGGG")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidHexColor, err)
}

func TestNewHexColor_EmptyString(t *testing.T) {
	_, err := NewHexColor("")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidHexColor, err)
}

func TestHexColor_Equals(t *testing.T) {
	color1, _ := NewHexColor("#FF5733")
	color2, _ := NewHexColor("#FF5733")
	color3, _ := NewHexColor("#00FF00")

	assert.True(t, color1.Equals(color2))
	assert.False(t, color1.Equals(color3))
}

func TestHexColor_EqualsCaseSensitive(t *testing.T) {
	color1, _ := NewHexColor("#FFFFFF")
	color2, _ := NewHexColor("#ffffff")

	assert.False(t, color1.Equals(color2))
}
