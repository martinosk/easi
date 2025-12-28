package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"errors"
	"strings"
)

var (
	ErrInvalidCapabilityStatus = errors.New("invalid capability status: must be Active, Planned, or Deprecated")
)

type CapabilityStatus string

const (
	StatusActive     CapabilityStatus = "Active"
	StatusPlanned    CapabilityStatus = "Planned"
	StatusDeprecated CapabilityStatus = "Deprecated"
)

func NewCapabilityStatus(value string) (CapabilityStatus, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return StatusActive, nil
	}

	switch CapabilityStatus(normalized) {
	case StatusActive, StatusPlanned, StatusDeprecated:
		return CapabilityStatus(normalized), nil
	default:
		return "", ErrInvalidCapabilityStatus
	}
}

func (c CapabilityStatus) Value() string {
	return string(c)
}

func (c CapabilityStatus) Equals(other domain.ValueObject) bool {
	if otherStatus, ok := other.(CapabilityStatus); ok {
		return c == otherStatus
	}
	return false
}

func (c CapabilityStatus) String() string {
	return string(c)
}
