package valueobjects

import (
	"errors"
	"regexp"

	"easi/backend/internal/shared/eventsourcing"
)

var (
	ErrInvalidHexColor = errors.New("invalid hex color: must be in format #RRGGBB")
	hexColorRegex      = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
)

type HexColor struct {
	value string
}

func NewHexColor(value string) (HexColor, error) {
	if !hexColorRegex.MatchString(value) {
		return HexColor{}, ErrInvalidHexColor
	}
	return HexColor{value: value}, nil
}

func (h HexColor) Value() string {
	return h.value
}

func (h HexColor) Equals(other domain.ValueObject) bool {
	if otherColor, ok := other.(HexColor); ok {
		return h.value == otherColor.value
	}
	return false
}

func (h HexColor) String() string {
	return h.value
}
