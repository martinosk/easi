package valueobjects

import (
	"errors"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrFitRationaleTooLong = errors.New("fit rationale cannot exceed 500 characters")

type FitRationale struct {
	value string
}

func NewFitRationale(value string) (FitRationale, error) {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) > 500 {
		return FitRationale{}, ErrFitRationaleTooLong
	}
	return FitRationale{value: trimmed}, nil
}

func (f FitRationale) Value() string {
	return f.value
}

func (f FitRationale) IsEmpty() bool {
	return f.value == ""
}

func (f FitRationale) Equals(other domain.ValueObject) bool {
	if otherRationale, ok := other.(FitRationale); ok {
		return f.value == otherRationale.value
	}
	return false
}

func (f FitRationale) String() string {
	return f.value
}
