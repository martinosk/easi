package valueobjects

import (
	"errors"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrPillarNameEmpty   = errors.New("pillar name cannot be empty or whitespace only")
	ErrPillarNameTooLong = errors.New("pillar name cannot exceed 100 characters")
)

type PillarName struct {
	value string
}

func NewPillarName(value string) (PillarName, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return PillarName{}, ErrPillarNameEmpty
	}
	if len(trimmed) > 100 {
		return PillarName{}, ErrPillarNameTooLong
	}
	return PillarName{value: trimmed}, nil
}

func (p PillarName) Value() string {
	return p.value
}

func (p PillarName) EqualsIgnoreCase(other PillarName) bool {
	return strings.EqualFold(p.value, other.value)
}

func (p PillarName) Equals(other domain.ValueObject) bool {
	if otherName, ok := other.(PillarName); ok {
		return p.value == otherName.value
	}
	return false
}

func (p PillarName) String() string {
	return p.value
}
