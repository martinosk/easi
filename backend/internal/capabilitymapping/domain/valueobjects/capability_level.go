package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"errors"
	"strings"
)

var (
	ErrInvalidCapabilityLevel = errors.New("invalid capability level: must be L1, L2, L3, or L4")
)

type CapabilityLevel string

const (
	LevelL1 CapabilityLevel = "L1"
	LevelL2 CapabilityLevel = "L2"
	LevelL3 CapabilityLevel = "L3"
	LevelL4 CapabilityLevel = "L4"
)

func NewCapabilityLevel(value string) (CapabilityLevel, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))

	switch CapabilityLevel(normalized) {
	case LevelL1, LevelL2, LevelL3, LevelL4:
		return CapabilityLevel(normalized), nil
	default:
		return "", ErrInvalidCapabilityLevel
	}
}

func (c CapabilityLevel) Value() string {
	return string(c)
}

func (c CapabilityLevel) Equals(other domain.ValueObject) bool {
	if otherLevel, ok := other.(CapabilityLevel); ok {
		return c == otherLevel
	}
	return false
}

func (c CapabilityLevel) String() string {
	return string(c)
}

func (c CapabilityLevel) IsValid() bool {
	return c == LevelL1 || c == LevelL2 || c == LevelL3 || c == LevelL4
}

func (c CapabilityLevel) NumericValue() int {
	switch c {
	case LevelL1:
		return 1
	case LevelL2:
		return 2
	case LevelL3:
		return 3
	case LevelL4:
		return 4
	default:
		return 0
	}
}
