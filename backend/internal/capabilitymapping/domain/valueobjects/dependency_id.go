package valueobjects

import (
	"easi/backend/internal/shared/domain"

	"github.com/google/uuid"
)

type DependencyID struct {
	value string
}

func NewDependencyID() DependencyID {
	return DependencyID{value: uuid.New().String()}
}

func NewDependencyIDFromString(value string) (DependencyID, error) {
	if value == "" {
		return DependencyID{}, domain.ErrEmptyValue
	}

	if _, err := uuid.Parse(value); err != nil {
		return DependencyID{}, domain.ErrInvalidValue
	}

	return DependencyID{value: value}, nil
}

func (d DependencyID) Value() string {
	return d.value
}

func (d DependencyID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(DependencyID); ok {
		return d.value == otherID.value
	}
	return false
}

func (d DependencyID) String() string {
	return d.value
}
