package valueobjects

import (
	"easi/backend/internal/shared/eventsourcing"
	"errors"
	"strings"
)

var (
	ErrTenantNameEmpty   = errors.New("tenant name cannot be empty")
	ErrTenantNameTooLong = errors.New("tenant name cannot exceed 255 characters")
)

type TenantName struct {
	value string
}

func NewTenantName(value string) (TenantName, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return TenantName{}, ErrTenantNameEmpty
	}
	if len(trimmed) > 255 {
		return TenantName{}, ErrTenantNameTooLong
	}
	return TenantName{value: trimmed}, nil
}

func (t TenantName) Value() string {
	return t.value
}

func (t TenantName) String() string {
	return t.value
}

func (t TenantName) Equals(other domain.ValueObject) bool {
	if otherName, ok := other.(TenantName); ok {
		return t.value == otherName.value
	}
	return false
}
