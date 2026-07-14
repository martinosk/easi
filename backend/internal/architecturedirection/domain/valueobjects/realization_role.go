package valueobjects

import (
	"errors"

	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrInvalidRealizationRole = errors.New("realization role must be one of standard, legacy")

const (
	RealizationRoleStandard = "standard"
	RealizationRoleLegacy   = "legacy"
)

type RealizationRole struct {
	value string
}

func NewRealizationRole(value string) (RealizationRole, error) {
	switch value {
	case RealizationRoleStandard, RealizationRoleLegacy:
		return RealizationRole{value: value}, nil
	default:
		return RealizationRole{}, ErrInvalidRealizationRole
	}
}

func (r RealizationRole) Value() string { return r.value }

func (r RealizationRole) Equals(other domain.ValueObject) bool {
	if otherRole, ok := other.(RealizationRole); ok {
		return r.value == otherRole.value
	}
	return false
}
