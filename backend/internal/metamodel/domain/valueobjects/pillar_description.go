package valueobjects

import (
	"errors"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrPillarDescriptionTooLong = errors.New("pillar description cannot exceed 500 characters")
)

type PillarDescription struct {
	value string
}

func NewPillarDescription(value string) (PillarDescription, error) {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) > 500 {
		return PillarDescription{}, ErrPillarDescriptionTooLong
	}
	return PillarDescription{value: trimmed}, nil
}

func (p PillarDescription) Value() string {
	return p.value
}

func (p PillarDescription) IsEmpty() bool {
	return p.value == ""
}

func (p PillarDescription) Equals(other domain.ValueObject) bool {
	if otherDesc, ok := other.(PillarDescription); ok {
		return p.value == otherDesc.value
	}
	return false
}

func (p PillarDescription) String() string {
	return p.value
}
