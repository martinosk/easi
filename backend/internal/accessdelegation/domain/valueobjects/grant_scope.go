package valueobjects

import (
	"errors"

	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrInvalidGrantScope = errors.New("invalid grant scope")

type GrantScope string

const (
	GrantScopeWrite GrantScope = "write"
)

func NewGrantScope(s string) (GrantScope, error) {
	switch s {
	case "write":
		return GrantScopeWrite, nil
	default:
		return "", ErrInvalidGrantScope
	}
}

func (gs GrantScope) String() string {
	return string(gs)
}

func (gs GrantScope) Equals(other domain.ValueObject) bool {
	otherScope, ok := other.(GrantScope)
	if !ok {
		return false
	}
	return gs == otherScope
}
