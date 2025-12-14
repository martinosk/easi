package valueobjects

import (
	"errors"
	"strings"

	"easi/backend/internal/shared/domain"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	ErrInvalidRealizationLevel = errors.New("invalid realization level: must be Full, Partial, or Planned")
)

type RealizationLevel string

const (
	RealizationFull    RealizationLevel = "Full"
	RealizationPartial RealizationLevel = "Partial"
	RealizationPlanned RealizationLevel = "Planned"
)

func NewRealizationLevel(value string) (RealizationLevel, error) {
	caser := cases.Title(language.English)
	normalized := caser.String(strings.ToLower(strings.TrimSpace(value)))

	switch RealizationLevel(normalized) {
	case RealizationFull, RealizationPartial, RealizationPlanned:
		return RealizationLevel(normalized), nil
	default:
		return "", ErrInvalidRealizationLevel
	}
}

func (r RealizationLevel) Value() string {
	return string(r)
}

func (r RealizationLevel) Equals(other domain.ValueObject) bool {
	if otherLevel, ok := other.(RealizationLevel); ok {
		return r == otherLevel
	}
	return false
}

func (r RealizationLevel) String() string {
	return string(r)
}

func (r RealizationLevel) IsValid() bool {
	return r == RealizationFull || r == RealizationPartial || r == RealizationPlanned
}
