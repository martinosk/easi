package valueobjects

import (
	"errors"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
)

const (
	FitTypeTechnical  = "TECHNICAL"
	FitTypeFunctional = "FUNCTIONAL"
)

var ErrInvalidFitType = errors.New("fit type must be TECHNICAL, FUNCTIONAL, or empty")

type FitType struct {
	value string
}

func NewFitType(value string) (FitType, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	if !isValidFitType(normalized) {
		return FitType{}, ErrInvalidFitType
	}
	return FitType{value: normalized}, nil
}

func isValidFitType(value string) bool {
	return value == "" || value == FitTypeTechnical || value == FitTypeFunctional
}

func (f FitType) Value() string {
	return f.value
}

func (f FitType) IsEmpty() bool {
	return f.value == ""
}

func (f FitType) IsTechnical() bool {
	return f.value == FitTypeTechnical
}

func (f FitType) IsFunctional() bool {
	return f.value == FitTypeFunctional
}

func (f FitType) Equals(other domain.ValueObject) bool {
	if otherFitType, ok := other.(FitType); ok {
		return f.value == otherFitType.value
	}
	return false
}

func (f FitType) String() string {
	return f.value
}
