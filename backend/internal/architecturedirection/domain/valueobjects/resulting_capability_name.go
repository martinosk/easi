package valueobjects

import (
	"errors"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
)

const MaxResultingCapabilityNameLength = 200

var (
	ErrResultingCapabilityNameRequired = errors.New("resulting name is required")
	ErrResultingCapabilityNameTooLong  = errors.New("resulting name exceeds maximum length of 200 characters")
)

type ResultingCapabilityName struct {
	value string
}

func NewResultingCapabilityName(value string) (ResultingCapabilityName, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ResultingCapabilityName{}, ErrResultingCapabilityNameRequired
	}
	if len(trimmed) > MaxResultingCapabilityNameLength {
		return ResultingCapabilityName{}, ErrResultingCapabilityNameTooLong
	}
	return ResultingCapabilityName{value: trimmed}, nil
}

func (n ResultingCapabilityName) Value() string { return n.value }

func (n ResultingCapabilityName) Equals(other domain.ValueObject) bool {
	if o, ok := other.(ResultingCapabilityName); ok {
		return n.value == o.value
	}
	return false
}
