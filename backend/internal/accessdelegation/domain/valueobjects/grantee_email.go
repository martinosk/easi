package valueobjects

import (
	"errors"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrGranteeEmailEmpty   = errors.New("grantee email must not be empty")
	ErrGranteeEmailInvalid = errors.New("grantee email format is invalid")
)

type GranteeEmail struct {
	value string
}

func NewGranteeEmail(email string) (GranteeEmail, error) {
	trimmed := strings.TrimSpace(email)
	if trimmed == "" {
		return GranteeEmail{}, ErrGranteeEmailEmpty
	}
	if !strings.Contains(trimmed, "@") {
		return GranteeEmail{}, ErrGranteeEmailInvalid
	}
	return GranteeEmail{value: strings.ToLower(trimmed)}, nil
}

func (e GranteeEmail) Value() string { return e.value }

func (e GranteeEmail) Equals(other domain.ValueObject) bool {
	otherEmail, ok := other.(GranteeEmail)
	if !ok {
		return false
	}
	return e.value == otherEmail.value
}
