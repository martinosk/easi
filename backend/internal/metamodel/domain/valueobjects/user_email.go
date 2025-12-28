package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"errors"
	"regexp"
	"strings"
)

var (
	ErrUserEmailEmpty   = errors.New("user email cannot be empty")
	ErrUserEmailInvalid = errors.New("user email must be a valid email address")
	ErrUserEmailTooLong = errors.New("user email cannot exceed 255 characters")

	emailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

type UserEmail struct {
	value string
}

func NewUserEmail(value string) (UserEmail, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return UserEmail{}, ErrUserEmailEmpty
	}
	if len(trimmed) > 255 {
		return UserEmail{}, ErrUserEmailTooLong
	}
	if !emailPattern.MatchString(trimmed) {
		return UserEmail{}, ErrUserEmailInvalid
	}
	return UserEmail{value: trimmed}, nil
}

func (u UserEmail) Value() string {
	return u.value
}

func (u UserEmail) Equals(other domain.ValueObject) bool {
	if otherEmail, ok := other.(UserEmail); ok {
		return u.value == otherEmail.value
	}
	return false
}

func (u UserEmail) String() string {
	return u.value
}
