package valueobjects

import (
	"errors"
	"net/url"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
)

const MaxURLLength = 2048

var (
	ErrURLEmpty   = errors.New("url cannot be empty")
	ErrURLInvalid = errors.New("url must be an absolute http or https url")
	ErrURLTooLong = errors.New("url exceeds maximum length of 2048 characters")
)

type URL struct {
	value string
}

func NewURL(value string) (URL, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return URL{}, ErrURLEmpty
	}
	if len(trimmed) > MaxURLLength {
		return URL{}, ErrURLTooLong
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || !isAbsoluteHTTP(parsed) {
		return URL{}, ErrURLInvalid
	}
	return URL{value: trimmed}, nil
}

func isAbsoluteHTTP(parsed *url.URL) bool {
	if parsed.Host == "" {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func (u URL) Value() string {
	return u.value
}

func (u URL) Equals(other domain.ValueObject) bool {
	if o, ok := other.(URL); ok {
		return u.value == o.value
	}
	return false
}

func (u URL) String() string {
	return u.value
}
