package valueobjects

import (
	"errors"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
)

const MaxEnterpriseCapabilityNameLength = 200

var (
	ErrEnterpriseCapabilityNameEmpty   = errors.New("enterprise capability name cannot be empty or whitespace only")
	ErrEnterpriseCapabilityNameTooLong = errors.New("enterprise capability name cannot exceed 200 characters")
)

type EnterpriseCapabilityName struct {
	value string
}

func NewEnterpriseCapabilityName(value string) (EnterpriseCapabilityName, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return EnterpriseCapabilityName{}, ErrEnterpriseCapabilityNameEmpty
	}
	if len(trimmed) > MaxEnterpriseCapabilityNameLength {
		return EnterpriseCapabilityName{}, ErrEnterpriseCapabilityNameTooLong
	}
	return EnterpriseCapabilityName{value: trimmed}, nil
}

func (e EnterpriseCapabilityName) Value() string {
	return e.value
}

func (e EnterpriseCapabilityName) String() string {
	return e.value
}

func (e EnterpriseCapabilityName) EqualsIgnoreCase(other string) bool {
	return strings.EqualFold(e.value, other)
}

func (e EnterpriseCapabilityName) Equals(other domain.ValueObject) bool {
	if otherName, ok := other.(EnterpriseCapabilityName); ok {
		return e.value == otherName.value
	}
	return false
}
