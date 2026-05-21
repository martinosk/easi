package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type DirectionID struct {
	sharedvo.UUIDValue
}

func NewDirectionID() DirectionID {
	return DirectionID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewDirectionIDFromString(value string) (DirectionID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return DirectionID{}, err
	}
	return DirectionID{UUIDValue: uuidValue}, nil
}

func (d DirectionID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(DirectionID); ok {
		return d.EqualsValue(otherID.UUIDValue)
	}
	return false
}
