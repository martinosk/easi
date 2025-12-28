package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"errors"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	ErrInvalidDependencyType = errors.New("invalid dependency type: must be Requires, Enables, or Supports")
)

type DependencyType string

const (
	DependencyRequires DependencyType = "Requires"
	DependencyEnables  DependencyType = "Enables"
	DependencySupports DependencyType = "Supports"
)

func NewDependencyType(value string) (DependencyType, error) {
	caser := cases.Title(language.English)
	normalized := caser.String(strings.ToLower(strings.TrimSpace(value)))

	switch DependencyType(normalized) {
	case DependencyRequires, DependencyEnables, DependencySupports:
		return DependencyType(normalized), nil
	default:
		return "", ErrInvalidDependencyType
	}
}

func (d DependencyType) Value() string {
	return string(d)
}

func (d DependencyType) Equals(other domain.ValueObject) bool {
	if otherType, ok := other.(DependencyType); ok {
		return d == otherType
	}
	return false
}

func (d DependencyType) String() string {
	return string(d)
}

func (d DependencyType) IsValid() bool {
	return d == DependencyRequires || d == DependencyEnables || d == DependencySupports
}
