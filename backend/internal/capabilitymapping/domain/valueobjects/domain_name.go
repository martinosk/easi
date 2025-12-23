package valueobjects

import (
	"errors"
	"strings"

	"easi/backend/internal/shared/eventsourcing"
)

var (
	ErrDomainNameEmpty   = errors.New("domain name cannot be empty or whitespace only")
	ErrDomainNameTooLong = errors.New("domain name cannot exceed 100 characters")
)

type DomainName struct {
	value string
}

func NewDomainName(value string) (DomainName, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return DomainName{}, ErrDomainNameEmpty
	}

	if len(trimmed) > 100 {
		return DomainName{}, ErrDomainNameTooLong
	}

	return DomainName{value: trimmed}, nil
}

func (d DomainName) Value() string {
	return d.value
}

func (d DomainName) Equals(other domain.ValueObject) bool {
	if otherName, ok := other.(DomainName); ok {
		return d.value == otherName.value
	}
	return false
}

func (d DomainName) String() string {
	return d.value
}
