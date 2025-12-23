package valueobjects

import (
	"errors"
	"strings"

	"easi/backend/internal/shared/eventsourcing"
)

var (
	ErrEmptyContextType   = errors.New("context type cannot be empty")
	ErrInvalidContextType = errors.New("invalid context type: must be architecture-canvas or business-domain-grid")
)

type LayoutContextType string

const (
	ContextTypeArchitectureCanvas LayoutContextType = "architecture-canvas"
	ContextTypeBusinessDomainGrid LayoutContextType = "business-domain-grid"
)

func NewLayoutContextType(value string) (LayoutContextType, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "", ErrEmptyContextType
	}

	switch LayoutContextType(normalized) {
	case ContextTypeArchitectureCanvas, ContextTypeBusinessDomainGrid:
		return LayoutContextType(normalized), nil
	default:
		return "", ErrInvalidContextType
	}
}

func (l LayoutContextType) Value() string {
	return string(l)
}

func (l LayoutContextType) Equals(other domain.ValueObject) bool {
	if otherType, ok := other.(LayoutContextType); ok {
		return l == otherType
	}
	return false
}

func (l LayoutContextType) String() string {
	return string(l)
}
