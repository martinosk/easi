package valueobjects

import (
	"errors"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrStageNameEmpty   = errors.New("stage name cannot be empty or whitespace only")
	ErrStageNameTooLong = errors.New("stage name cannot exceed 100 characters")
)

type StageName struct {
	value string
}

func NewStageName(value string) (StageName, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return StageName{}, ErrStageNameEmpty
	}

	if len(trimmed) > 100 {
		return StageName{}, ErrStageNameTooLong
	}

	return StageName{value: trimmed}, nil
}

func (n StageName) Value() string {
	return n.value
}

func (n StageName) Equals(other domain.ValueObject) bool {
	if otherName, ok := other.(StageName); ok {
		return n.value == otherName.value
	}
	return false
}

func (n StageName) String() string {
	return n.value
}
