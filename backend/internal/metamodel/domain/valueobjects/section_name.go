package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"errors"
	"strings"
)

var (
	ErrSectionNameEmpty   = errors.New("section name cannot be empty or whitespace only")
	ErrSectionNameTooLong = errors.New("section name cannot exceed 50 characters")
)

type SectionName struct {
	value string
}

func NewSectionName(value string) (SectionName, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return SectionName{}, ErrSectionNameEmpty
	}
	if len(trimmed) > 50 {
		return SectionName{}, ErrSectionNameTooLong
	}
	return SectionName{value: trimmed}, nil
}

func (s SectionName) Value() string {
	return s.value
}

func (s SectionName) Equals(other domain.ValueObject) bool {
	if otherName, ok := other.(SectionName); ok {
		return s.value == otherName.value
	}
	return false
}

func (s SectionName) String() string {
	return s.value
}
