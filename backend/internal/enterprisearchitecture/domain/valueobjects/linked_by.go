package valueobjects

import (
	"errors"
	"regexp"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
)

const SystemLinkedBy = "system"

var (
	ErrLinkedByEmpty   = errors.New("linkedBy cannot be empty")
	ErrLinkedByInvalid = errors.New("linkedBy must be a valid email address or 'system'")
	ErrLinkedByTooLong = errors.New("linkedBy cannot exceed 255 characters")

	linkedByEmailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

type LinkedBy struct {
	value string
}

func NewLinkedBy(value string) (LinkedBy, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return LinkedBy{}, ErrLinkedByEmpty
	}
	if len(trimmed) > 255 {
		return LinkedBy{}, ErrLinkedByTooLong
	}
	if strings.EqualFold(trimmed, SystemLinkedBy) {
		return LinkedBy{value: SystemLinkedBy}, nil
	}
	if !linkedByEmailPattern.MatchString(trimmed) {
		return LinkedBy{}, ErrLinkedByInvalid
	}
	return LinkedBy{value: strings.ToLower(trimmed)}, nil
}

func MustNewLinkedBy(value string) LinkedBy {
	lb, err := NewLinkedBy(value)
	if err != nil {
		panic(err)
	}
	return lb
}

func (l LinkedBy) Value() string {
	return l.value
}

func (l LinkedBy) String() string {
	return l.value
}

func (l LinkedBy) IsSystem() bool {
	return l.value == SystemLinkedBy
}

func (l LinkedBy) Equals(other domain.ValueObject) bool {
	if otherLB, ok := other.(LinkedBy); ok {
		return l.value == otherLB.value
	}
	return false
}
