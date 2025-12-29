package valueobjects

import (
	"errors"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
)

const MaxRationaleLength = 500

var ErrRationaleTooLong = errors.New("rationale cannot exceed 500 characters")

type Rationale struct {
	value string
}

func NewRationale(value string) (Rationale, error) {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) > MaxRationaleLength {
		return Rationale{}, ErrRationaleTooLong
	}
	return Rationale{value: trimmed}, nil
}

func EmptyRationale() Rationale {
	return Rationale{value: ""}
}

func (r Rationale) Value() string {
	return r.value
}

func (r Rationale) IsEmpty() bool {
	return r.value == ""
}

func (r Rationale) Equals(other domain.ValueObject) bool {
	if otherRationale, ok := other.(Rationale); ok {
		return r.value == otherRationale.value
	}
	return false
}
