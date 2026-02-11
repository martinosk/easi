package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type ValueStreamID struct {
	sharedvo.UUIDValue
}

func NewValueStreamID() ValueStreamID {
	return ValueStreamID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewValueStreamIDFromString(value string) (ValueStreamID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return ValueStreamID{}, err
	}
	return ValueStreamID{UUIDValue: uuidValue}, nil
}

func (v ValueStreamID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(ValueStreamID); ok {
		return v.UUIDValue.EqualsValue(otherID.UUIDValue)
	}
	return false
}
