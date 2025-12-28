package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type CapabilityID struct {
	sharedvo.UUIDValue
}

func NewCapabilityID() CapabilityID {
	return CapabilityID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewCapabilityIDFromString(value string) (CapabilityID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return CapabilityID{}, err
	}
	return CapabilityID{UUIDValue: uuidValue}, nil
}

func (c CapabilityID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(CapabilityID); ok {
		return c.UUIDValue.EqualsValue(otherID.UUIDValue)
	}
	return false
}
