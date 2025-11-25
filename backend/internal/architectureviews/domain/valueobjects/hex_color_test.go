package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHexColor_ValidUppercase(t *testing.T) {
	hexColor, err := NewHexColor("#FFFFFF")
	assert.NoError(t, err)
	assert.Equal(t, "#FFFFFF", hexColor.Value())
}

func TestNewHexColor_ValidLowercase(t *testing.T) {
	hexColor, err := NewHexColor("#ffffff")
	assert.NoError(t, err)
	assert.Equal(t, "#ffffff", hexColor.Value())
}

func TestNewHexColor_ValidMixedCase(t *testing.T) {
	hexColor, err := NewHexColor("#FfA5b3")
	assert.NoError(t, err)
	assert.Equal(t, "#FfA5b3", hexColor.Value())
}

func TestNewHexColor_ValidBlack(t *testing.T) {
	hexColor, err := NewHexColor("#000000")
	assert.NoError(t, err)
	assert.Equal(t, "#000000", hexColor.Value())
}

func TestNewHexColor_ValidRed(t *testing.T) {
	hexColor, err := NewHexColor("#FF0000")
	assert.NoError(t, err)
	assert.Equal(t, "#FF0000", hexColor.Value())
}

func TestNewHexColor_ValidGreen(t *testing.T) {
	hexColor, err := NewHexColor("#00FF00")
	assert.NoError(t, err)
	assert.Equal(t, "#00FF00", hexColor.Value())
}

func TestNewHexColor_ValidBlue(t *testing.T) {
	hexColor, err := NewHexColor("#0000FF")
	assert.NoError(t, err)
	assert.Equal(t, "#0000FF", hexColor.Value())
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

func TestNewHexColor_TooLong(t *testing.T) {
	_, err := NewHexColor("#FFFFFF00")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidHexColor, err)
}

func TestNewHexColor_InvalidCharactersG(t *testing.T) {
	_, err := NewHexColor("#GGGGGG")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidHexColor, err)
}

func TestNewHexColor_InvalidCharactersZ(t *testing.T) {
	_, err := NewHexColor("#ZZZZZZ")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidHexColor, err)
}

func TestNewHexColor_InvalidCharactersSpecial(t *testing.T) {
	_, err := NewHexColor("#FF@FF!")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidHexColor, err)
}

func TestNewHexColor_EmptyString(t *testing.T) {
	_, err := NewHexColor("")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidHexColor, err)
}

func TestNewHexColor_OnlyHash(t *testing.T) {
	_, err := NewHexColor("#")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidHexColor, err)
}

func TestNewHexColor_Whitespace(t *testing.T) {
	_, err := NewHexColor("#FFF FFF")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidHexColor, err)
}

func TestNewHexColor_LeadingWhitespace(t *testing.T) {
	_, err := NewHexColor(" #FFFFFF")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidHexColor, err)
}

func TestNewHexColor_TrailingWhitespace(t *testing.T) {
	_, err := NewHexColor("#FFFFFF ")
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

func TestHexColor_String(t *testing.T) {
	hexColor, _ := NewHexColor("#FF5733")
	assert.Equal(t, "#FF5733", hexColor.String())
}
