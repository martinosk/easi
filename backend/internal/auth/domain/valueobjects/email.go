package valueobjects

import (
	"errors"
	"net/mail"
	"strings"

	"easi/backend/internal/shared/eventsourcing"
)

var (
	ErrInvalidEmailFormat = errors.New("invalid email format")
	ErrEmailEmpty         = errors.New("email cannot be empty")
)

type Email struct {
	value  string
	domain string
}

func NewEmail(value string) (Email, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return Email{}, ErrEmailEmpty
	}

	addr, err := mail.ParseAddress(value)
	if err != nil {
		return Email{}, ErrInvalidEmailFormat
	}

	normalized := strings.ToLower(addr.Address)
	localPart, domain, valid := parseEmailParts(normalized)
	if !valid {
		return Email{}, ErrInvalidEmailFormat
	}

	return Email{
		value:  localPart + "@" + domain,
		domain: domain,
	}, nil
}

func parseEmailParts(email string) (localPart, domain string, valid bool) {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "", "", false
	}
	if parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func (e Email) Value() string {
	return e.value
}

func (e Email) Domain() string {
	return e.domain
}

func (e Email) Equals(other domain.ValueObject) bool {
	otherEmail, ok := other.(Email)
	if !ok {
		return false
	}
	return e.value == otherEmail.value
}
