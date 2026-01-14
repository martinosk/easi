package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"errors"
	"strings"
)

var (
	ErrExpertNameEmpty = errors.New("expert name cannot be empty or whitespace only")
)

type ExpertName struct {
	value string
}

func NewExpertName(value string) (ExpertName, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ExpertName{}, ErrExpertNameEmpty
	}
	return ExpertName{value: trimmed}, nil
}

func (e ExpertName) Value() string {
	return e.value
}

func (e ExpertName) Equals(other domain.ValueObject) bool {
	if otherName, ok := other.(ExpertName); ok {
		return e.value == otherName.value
	}
	return false
}

func (e ExpertName) String() string {
	return e.value
}
