package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"errors"
	"strings"
)

var (
	ErrValueStreamNameEmpty   = errors.New("value stream name cannot be empty or whitespace only")
	ErrValueStreamNameTooLong = errors.New("value stream name cannot exceed 100 characters")
)

type ValueStreamName struct {
	value string
}

func NewValueStreamName(value string) (ValueStreamName, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ValueStreamName{}, ErrValueStreamNameEmpty
	}

	if len(trimmed) > 100 {
		return ValueStreamName{}, ErrValueStreamNameTooLong
	}

	return ValueStreamName{value: trimmed}, nil
}

func MustNewValueStreamName(value string) ValueStreamName {
	name, err := NewValueStreamName(value)
	if err != nil {
		panic(err)
	}
	return name
}

func (n ValueStreamName) Value() string {
	return n.value
}

func (n ValueStreamName) Equals(other domain.ValueObject) bool {
	if otherName, ok := other.(ValueStreamName); ok {
		return n.value == otherName.value
	}
	return false
}

func (n ValueStreamName) String() string {
	return n.value
}
