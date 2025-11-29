package valueobjects

import (
	"errors"
	"strings"

	"easi/backend/internal/shared/domain"
	"github.com/google/uuid"
)

var (
	ErrBusinessDomainIDMissingPrefix = errors.New("business domain ID must start with 'bd-' prefix")
)

type BusinessDomainID struct {
	value string
}

func NewBusinessDomainID() BusinessDomainID {
	return BusinessDomainID{value: "bd-" + uuid.New().String()}
}

func NewBusinessDomainIDFromString(value string) (BusinessDomainID, error) {
	if value == "" {
		return BusinessDomainID{}, domain.ErrEmptyValue
	}

	if !strings.HasPrefix(value, "bd-") {
		return BusinessDomainID{}, ErrBusinessDomainIDMissingPrefix
	}

	guidPart := strings.TrimPrefix(value, "bd-")
	if _, err := uuid.Parse(guidPart); err != nil {
		return BusinessDomainID{}, domain.ErrInvalidValue
	}

	return BusinessDomainID{value: value}, nil
}

func (b BusinessDomainID) Value() string {
	return b.value
}

func (b BusinessDomainID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(BusinessDomainID); ok {
		return b.value == otherID.value
	}
	return false
}

func (b BusinessDomainID) String() string {
	return b.value
}
