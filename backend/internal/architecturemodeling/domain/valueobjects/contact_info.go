package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"errors"
	"strings"
)

var (
	ErrContactInfoEmpty = errors.New("contact info cannot be empty or whitespace only")
)

type ContactInfo struct {
	value string
}

func NewContactInfo(value string) (ContactInfo, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ContactInfo{}, ErrContactInfoEmpty
	}
	return ContactInfo{value: trimmed}, nil
}

func (c ContactInfo) Value() string {
	return c.value
}

func (c ContactInfo) Equals(other domain.ValueObject) bool {
	if otherInfo, ok := other.(ContactInfo); ok {
		return c.value == otherInfo.value
	}
	return false
}

func (c ContactInfo) String() string {
	return c.value
}
