package valueobjects

import (
	"errors"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrStewardIDRequired = errors.New("steward user id is required")
	ErrStewardNotFound   = errors.New("steward is not a user of this tenant")
	ErrStewardDisabled   = errors.New("steward is a disabled user")
)

type UserStanding int

const (
	UserUnknown UserStanding = iota
	UserActive
	UserDisabled
)

type StewardRef struct {
	userID string
}

func NewStewardRef(userID string, standing UserStanding) (StewardRef, error) {
	trimmed := strings.TrimSpace(userID)
	if trimmed == "" {
		return StewardRef{}, ErrStewardIDRequired
	}
	switch standing {
	case UserActive:
		return StewardRef{userID: trimmed}, nil
	case UserDisabled:
		return StewardRef{}, ErrStewardDisabled
	default:
		return StewardRef{}, ErrStewardNotFound
	}
}

func RecordedStewardRef(userID string) StewardRef {
	return StewardRef{userID: userID}
}

func (s StewardRef) Value() string { return s.userID }

func (s StewardRef) Equals(other domain.ValueObject) bool {
	if otherRef, ok := other.(StewardRef); ok {
		return s.userID == otherRef.userID
	}
	return false
}
