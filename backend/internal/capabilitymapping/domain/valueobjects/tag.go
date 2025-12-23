package valueobjects

import (
	"errors"
	"strings"

	"easi/backend/internal/shared/eventsourcing"
)

var (
	ErrTagEmpty = errors.New("tag cannot be empty")
)

type Tag struct {
	value string
}

func NewTag(value string) (Tag, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return Tag{}, ErrTagEmpty
	}

	return Tag{value: trimmed}, nil
}

func (t Tag) Value() string {
	return t.value
}

func (t Tag) Equals(other domain.ValueObject) bool {
	if otherTag, ok := other.(Tag); ok {
		return t.value == otherTag.value
	}
	return false
}

func (t Tag) String() string {
	return t.value
}
