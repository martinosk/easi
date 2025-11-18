package valueobjects

import (
	"easi/backend/internal/shared/domain"
	"github.com/google/uuid"
)

type RealizationID struct {
	value string
}

func NewRealizationID() RealizationID {
	return RealizationID{value: uuid.New().String()}
}

func NewRealizationIDFromString(value string) (RealizationID, error) {
	if value == "" {
		return RealizationID{}, domain.ErrEmptyValue
	}

	if _, err := uuid.Parse(value); err != nil {
		return RealizationID{}, domain.ErrInvalidValue
	}

	return RealizationID{value: value}, nil
}

func (r RealizationID) Value() string {
	return r.value
}

func (r RealizationID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(RealizationID); ok {
		return r.value == otherID.value
	}
	return false
}

func (r RealizationID) String() string {
	return r.value
}
