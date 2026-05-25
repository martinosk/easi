package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type StandardApplicationID struct {
	sharedvo.UUIDValue
}

func NewStandardApplicationID() StandardApplicationID {
	return StandardApplicationID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewStandardApplicationIDFromString(value string) (StandardApplicationID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return StandardApplicationID{}, err
	}
	return StandardApplicationID{UUIDValue: uuidValue}, nil
}

func (s StandardApplicationID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(StandardApplicationID); ok {
		return s.EqualsValue(otherID.UUIDValue)
	}
	return false
}
