package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type ViewID struct {
	sharedvo.UUIDValue
}

func NewViewID() ViewID {
	return ViewID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewViewIDFromString(value string) (ViewID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return ViewID{}, err
	}
	return ViewID{UUIDValue: uuidValue}, nil
}

func (v ViewID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(ViewID); ok {
		return v.UUIDValue.EqualsValue(otherID.UUIDValue)
	}
	return false
}
