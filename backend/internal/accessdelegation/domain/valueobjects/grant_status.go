package valueobjects

import (
	"errors"

	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrInvalidGrantStatus = errors.New("invalid grant status")

type GrantStatus string

const (
	GrantStatusActive  GrantStatus = "active"
	GrantStatusRevoked GrantStatus = "revoked"
	GrantStatusExpired GrantStatus = "expired"
)

func NewGrantStatus(s string) (GrantStatus, error) {
	switch s {
	case "active":
		return GrantStatusActive, nil
	case "revoked":
		return GrantStatusRevoked, nil
	case "expired":
		return GrantStatusExpired, nil
	default:
		return "", ErrInvalidGrantStatus
	}
}

func (gs GrantStatus) String() string {
	return string(gs)
}

func (gs GrantStatus) IsActive() bool {
	return gs == GrantStatusActive
}

func (gs GrantStatus) IsRevoked() bool {
	return gs == GrantStatusRevoked
}

func (gs GrantStatus) IsExpired() bool {
	return gs == GrantStatusExpired
}

func (gs GrantStatus) Equals(other domain.ValueObject) bool {
	otherStatus, ok := other.(GrantStatus)
	if !ok {
		return false
	}
	return gs == otherStatus
}
