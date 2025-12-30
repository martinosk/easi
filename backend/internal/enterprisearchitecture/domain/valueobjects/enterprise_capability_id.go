package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type EnterpriseCapabilityID struct {
	sharedvo.UUIDValue
}

func NewEnterpriseCapabilityID() EnterpriseCapabilityID {
	return EnterpriseCapabilityID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewEnterpriseCapabilityIDFromString(value string) (EnterpriseCapabilityID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return EnterpriseCapabilityID{}, err
	}
	return EnterpriseCapabilityID{UUIDValue: uuidValue}, nil
}

func (e EnterpriseCapabilityID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(EnterpriseCapabilityID); ok {
		return e.UUIDValue.EqualsValue(otherID.UUIDValue)
	}
	return false
}
