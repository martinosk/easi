package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type AcquiredEntityID struct {
	sharedvo.UUIDValue
}

func NewAcquiredEntityID() AcquiredEntityID {
	return AcquiredEntityID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewAcquiredEntityIDFromString(value string) (AcquiredEntityID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return AcquiredEntityID{}, err
	}
	return AcquiredEntityID{UUIDValue: uuidValue}, nil
}

func (a AcquiredEntityID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(AcquiredEntityID); ok {
		return a.UUIDValue.EqualsValue(otherID.UUIDValue)
	}
	return false
}
