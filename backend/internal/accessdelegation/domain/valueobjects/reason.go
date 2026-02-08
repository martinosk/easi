package valueobjects

import (
	"errors"

	domain "easi/backend/internal/shared/eventsourcing"
)

const MaxReasonLength = 1000

var ErrReasonTooLong = errors.New("reason must not exceed 1000 characters")

type Reason struct {
	value string
}

func NewReason(s string) (Reason, error) {
	if len(s) > MaxReasonLength {
		return Reason{}, ErrReasonTooLong
	}
	return Reason{value: s}, nil
}

func (r Reason) Value() string { return r.value }

func (r Reason) Equals(other domain.ValueObject) bool {
	otherReason, ok := other.(Reason)
	if !ok {
		return false
	}
	return r.value == otherReason.value
}
