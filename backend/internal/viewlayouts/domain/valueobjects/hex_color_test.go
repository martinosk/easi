package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHexColor_Valid(t *testing.T) {
	tests := []string{
		"#000000",
		"#ffffff",
		"#FFFFFF",
		"#3b82f6",
		"#AB12CD",
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			color, err := NewHexColor(tt)
			assert.NoError(t, err)
			assert.Equal(t, tt, color.Value())
		})
	}
}

func TestNewHexColor_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"no hash", "000000"},
		{"short", "#fff"},
		{"too long", "#fffffff"},
		{"invalid chars", "#gggggg"},
		{"spaces", "# 00000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewHexColor(tt.input)
			assert.Error(t, err)
			assert.Equal(t, ErrInvalidHexColor, err)
		})
	}
}

func TestHexColor_String(t *testing.T) {
	color, _ := NewHexColor("#3b82f6")
	assert.Equal(t, "#3b82f6", color.String())
}

func TestHexColor_Equals(t *testing.T) {
	color1, _ := NewHexColor("#3b82f6")
	color2, _ := NewHexColor("#3b82f6")
	color3, _ := NewHexColor("#ffffff")

	assert.True(t, color1.Equals(color2))
	assert.False(t, color1.Equals(color3))
}
