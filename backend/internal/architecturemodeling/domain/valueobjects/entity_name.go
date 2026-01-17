package valueobjects

import (
	"errors"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
)

const MaxEntityNameLength = 100

var (
	ErrEntityNameEmpty   = errors.New("name cannot be empty or whitespace only")
	ErrEntityNameTooLong = errors.New("name exceeds maximum length of 100 characters")
)

type EntityName struct {
	value string
}

func NewEntityName(value string) (EntityName, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return EntityName{}, ErrEntityNameEmpty
	}
	if len(trimmed) > MaxEntityNameLength {
		return EntityName{}, ErrEntityNameTooLong
	}
	return EntityName{value: trimmed}, nil
}

func MustNewEntityName(value string) EntityName {
	name, err := NewEntityName(value)
	if err != nil {
		panic(err)
	}
	return name
}

func (e EntityName) Value() string {
	return e.value
}

func (e EntityName) Equals(other domain.ValueObject) bool {
	if otherName, ok := other.(EntityName); ok {
		return e.value == otherName.value
	}
	return false
}

func (e EntityName) String() string {
	return e.value
}
