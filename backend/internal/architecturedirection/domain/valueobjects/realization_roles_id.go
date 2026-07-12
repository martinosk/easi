package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type RealizationRolesID struct {
	sharedvo.UUIDValue
}

func NewRealizationRolesID() RealizationRolesID {
	return RealizationRolesID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewRealizationRolesIDFromString(value string) (RealizationRolesID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return RealizationRolesID{}, err
	}
	return RealizationRolesID{UUIDValue: uuidValue}, nil
}

func (i RealizationRolesID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(RealizationRolesID); ok {
		return i.EqualsValue(otherID.UUIDValue)
	}
	return false
}
