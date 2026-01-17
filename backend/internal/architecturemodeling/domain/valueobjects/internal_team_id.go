package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type InternalTeamID struct {
	sharedvo.UUIDValue
}

func NewInternalTeamID() InternalTeamID {
	return InternalTeamID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewInternalTeamIDFromString(value string) (InternalTeamID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return InternalTeamID{}, err
	}
	return InternalTeamID{UUIDValue: uuidValue}, nil
}

func (i InternalTeamID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(InternalTeamID); ok {
		return i.UUIDValue.EqualsValue(otherID.UUIDValue)
	}
	return false
}
