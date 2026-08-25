package valueobjects

import (
	"errors"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/google/uuid"
)

var (
	ErrEAOwnerNotUser   = errors.New("eaOwner must be a user id or a name that resolves to exactly one user")
	ErrEAOwnerAmbiguous = errors.New("eaOwner name matches more than one user")
)

type EAOwner struct {
	value string
}

func NewEAOwner(value string) (EAOwner, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return EAOwner{}, nil
	}
	if _, err := uuid.Parse(trimmed); err != nil {
		return EAOwner{}, ErrEAOwnerNotUser
	}
	return EAOwner{value: trimmed}, nil
}

func EAOwnerFromHistory(value string) EAOwner {
	return EAOwner{value: strings.TrimSpace(value)}
}

func (r EAOwner) Value() string {
	return r.value
}

func (r EAOwner) IsEmpty() bool {
	return r.value == ""
}

func (r EAOwner) Equals(other domain.ValueObject) bool {
	if otherRef, ok := other.(EAOwner); ok {
		return r.value == otherRef.value
	}
	return false
}

func (r EAOwner) String() string {
	return r.value
}
