package valueobjects

import (
	"errors"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrFitCriteriaTooLong = errors.New("fit criteria cannot exceed 500 characters")
)

type FitCriteria struct {
	value string
}

func NewFitCriteria(value string) (FitCriteria, error) {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) > 500 {
		return FitCriteria{}, ErrFitCriteriaTooLong
	}
	return FitCriteria{value: trimmed}, nil
}

func (f FitCriteria) Value() string {
	return f.value
}

func (f FitCriteria) IsEmpty() bool {
	return f.value == ""
}

func (f FitCriteria) Equals(other domain.ValueObject) bool {
	if otherCriteria, ok := other.(FitCriteria); ok {
		return f.value == otherCriteria.value
	}
	return false
}

func (f FitCriteria) String() string {
	return f.value
}
