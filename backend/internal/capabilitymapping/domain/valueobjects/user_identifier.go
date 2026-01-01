package valueobjects

import (
	"errors"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrUserIdentifierEmpty   = errors.New("user identifier cannot be empty")
	ErrUserIdentifierTooLong = errors.New("user identifier cannot exceed 255 characters")
)

const maxUserIdentifierLength = 255

type UserIdentifier struct {
	value string
}

func NewUserIdentifier(value string) (UserIdentifier, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return UserIdentifier{}, ErrUserIdentifierEmpty
	}
	if len(trimmed) > maxUserIdentifierLength {
		return UserIdentifier{}, ErrUserIdentifierTooLong
	}
	return UserIdentifier{value: trimmed}, nil
}

func (u UserIdentifier) Value() string {
	return u.value
}

func (u UserIdentifier) String() string {
	return u.value
}

func (u UserIdentifier) Equals(other domain.ValueObject) bool {
	if otherUser, ok := other.(UserIdentifier); ok {
		return u.value == otherUser.value
	}
	return false
}
