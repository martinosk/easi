package valueobjects

import (
	"errors"

	"easi/backend/internal/shared/domain"
)

var (
	ErrInvalidColorScheme = errors.New("invalid color scheme: must be 'maturity', 'classic', or 'custom'")
)

type ColorScheme struct {
	value string
}

func NewColorScheme(value string) (ColorScheme, error) {
	switch value {
	case "maturity", "classic", "custom":
		return ColorScheme{value: value}, nil
	default:
		return ColorScheme{}, ErrInvalidColorScheme
	}
}

func DefaultColorScheme() ColorScheme {
	return ColorScheme{value: "maturity"}
}

func (c ColorScheme) Value() string {
	return c.value
}

func (c ColorScheme) Equals(other domain.ValueObject) bool {
	if otherColorScheme, ok := other.(ColorScheme); ok {
		return c.value == otherColorScheme.value
	}
	return false
}

func (c ColorScheme) String() string {
	return c.value
}
