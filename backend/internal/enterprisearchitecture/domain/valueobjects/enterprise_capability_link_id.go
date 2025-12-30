package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type EnterpriseCapabilityLinkID struct {
	sharedvo.UUIDValue
}

func NewEnterpriseCapabilityLinkID() EnterpriseCapabilityLinkID {
	return EnterpriseCapabilityLinkID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewEnterpriseCapabilityLinkIDFromString(value string) (EnterpriseCapabilityLinkID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return EnterpriseCapabilityLinkID{}, err
	}
	return EnterpriseCapabilityLinkID{UUIDValue: uuidValue}, nil
}

func (e EnterpriseCapabilityLinkID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(EnterpriseCapabilityLinkID); ok {
		return e.UUIDValue.EqualsValue(otherID.UUIDValue)
	}
	return false
}
