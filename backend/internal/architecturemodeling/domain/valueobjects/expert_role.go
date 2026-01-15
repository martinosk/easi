package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"errors"
	"strings"
)

var (
	ErrExpertRoleEmpty = errors.New("expert role cannot be empty or whitespace only")
)

type ExpertRole struct {
	value string
}

func NewExpertRole(value string) (ExpertRole, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ExpertRole{}, ErrExpertRoleEmpty
	}
	return ExpertRole{value: trimmed}, nil
}

func (e ExpertRole) Value() string {
	return e.value
}

func (e ExpertRole) Equals(other domain.ValueObject) bool {
	if otherRole, ok := other.(ExpertRole); ok {
		return e.value == otherRole.value
	}
	return false
}

func (e ExpertRole) String() string {
	return e.value
}
