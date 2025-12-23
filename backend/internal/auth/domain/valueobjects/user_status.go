package valueobjects

import (
	"errors"

	"easi/backend/internal/shared/eventsourcing"
)

var ErrInvalidUserStatus = errors.New("invalid user status: must be active or disabled")

type UserStatus struct {
	value string
}

var (
	UserStatusActive   = UserStatus{value: "active"}
	UserStatusDisabled = UserStatus{value: "disabled"}
)

func UserStatusFromString(s string) (UserStatus, error) {
	switch s {
	case "active":
		return UserStatusActive, nil
	case "disabled":
		return UserStatusDisabled, nil
	default:
		return UserStatus{}, ErrInvalidUserStatus
	}
}

func (s UserStatus) String() string {
	return s.value
}

func (s UserStatus) IsActive() bool {
	return s.value == "active"
}

func (s UserStatus) Equals(other domain.ValueObject) bool {
	otherStatus, ok := other.(UserStatus)
	if !ok {
		return false
	}
	return s.value == otherStatus.value
}
