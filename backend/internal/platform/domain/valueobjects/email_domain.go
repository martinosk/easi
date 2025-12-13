package valueobjects

import (
	"easi/backend/internal/shared/domain"
	"errors"
	"regexp"
	"strings"
)

var (
	ErrEmailDomainEmpty     = errors.New("email domain cannot be empty")
	ErrInvalidEmailDomain   = errors.New("invalid email domain format")
	ErrEmailDomainListEmpty = errors.New("email domain list cannot be empty")
	ErrDuplicateEmailDomain = errors.New("duplicate email domain in list")

	emailDomainPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]*[a-z0-9]\.[a-z]{2,}$`)
)

type EmailDomain struct {
	value string
}

func NewEmailDomain(value string) (EmailDomain, error) {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	if trimmed == "" {
		return EmailDomain{}, ErrEmailDomainEmpty
	}

	if !isValidEmailDomain(trimmed) {
		return EmailDomain{}, ErrInvalidEmailDomain
	}

	return EmailDomain{value: trimmed}, nil
}

func isValidEmailDomain(domain string) bool {
	if len(domain) < 4 {
		return false
	}

	if strings.Contains(domain, "..") {
		return false
	}

	if strings.HasPrefix(domain, "-") || strings.HasSuffix(domain, "-") {
		return false
	}

	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}

	if strings.Contains(domain, "_") || strings.Contains(domain, " ") {
		return false
	}

	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return false
	}

	return emailDomainPattern.MatchString(domain)
}

func (e EmailDomain) Value() string {
	return e.value
}

func (e EmailDomain) String() string {
	return e.value
}

func (e EmailDomain) Equals(other domain.ValueObject) bool {
	if otherDomain, ok := other.(EmailDomain); ok {
		return e.value == otherDomain.value
	}
	return false
}

func NewEmailDomainList(domains []string) ([]EmailDomain, error) {
	if len(domains) == 0 {
		return nil, ErrEmailDomainListEmpty
	}

	seen := make(map[string]bool)
	result := make([]EmailDomain, 0, len(domains))

	for _, d := range domains {
		emailDomain, err := NewEmailDomain(d)
		if err != nil {
			return nil, err
		}

		if seen[emailDomain.value] {
			return nil, ErrDuplicateEmailDomain
		}
		seen[emailDomain.value] = true

		result = append(result, emailDomain)
	}

	return result, nil
}
