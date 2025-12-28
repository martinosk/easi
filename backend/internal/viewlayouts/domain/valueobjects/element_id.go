package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"errors"
	"strings"
)

var (
	ErrEmptyElementID = errors.New("element ID cannot be empty")
)

type ElementID struct {
	value string
}

func NewElementID(value string) (ElementID, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ElementID{}, ErrEmptyElementID
	}

	return ElementID{value: trimmed}, nil
}

func (e ElementID) Value() string {
	return e.value
}

func (e ElementID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(ElementID); ok {
		return e.value == otherID.value
	}
	return false
}

func (e ElementID) String() string {
	return e.value
}
