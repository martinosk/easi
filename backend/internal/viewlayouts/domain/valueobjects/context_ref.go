package valueobjects

import (
	"errors"
	"strings"

	"easi/backend/internal/shared/domain"
)

var (
	ErrEmptyContextRef = errors.New("context reference cannot be empty")
)

type ContextRef struct {
	value string
}

func NewContextRef(value string) (ContextRef, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ContextRef{}, ErrEmptyContextRef
	}

	return ContextRef{value: trimmed}, nil
}

func (c ContextRef) Value() string {
	return c.value
}

func (c ContextRef) Equals(other domain.ValueObject) bool {
	if otherRef, ok := other.(ContextRef); ok {
		return c.value == otherRef.value
	}
	return false
}

func (c ContextRef) String() string {
	return c.value
}
