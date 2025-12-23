package valueobjects

import (
	"easi/backend/internal/shared/eventsourcing"

	"github.com/google/uuid"
)

type UUIDValue struct {
	value string
}

func NewUUIDValue() UUIDValue {
	return UUIDValue{value: uuid.New().String()}
}

func NewUUIDValueFromString(value string) (UUIDValue, error) {
	if value == "" {
		return UUIDValue{}, domain.ErrEmptyValue
	}

	if _, err := uuid.Parse(value); err != nil {
		return UUIDValue{}, domain.ErrInvalidValue
	}

	return UUIDValue{value: value}, nil
}

func (u UUIDValue) Value() string {
	return u.value
}

func (u UUIDValue) String() string {
	return u.value
}

func (u UUIDValue) EqualsValue(other UUIDValue) bool {
	return u.value == other.value
}

func (u UUIDValue) Equals(other domain.ValueObject) bool {
	if otherUUID, ok := other.(UUIDValue); ok {
		return u.value == otherUUID.value
	}
	return false
}
