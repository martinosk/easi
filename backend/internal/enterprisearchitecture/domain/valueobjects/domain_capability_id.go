package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type DomainCapabilityID struct {
	sharedvo.UUIDValue
}

func NewDomainCapabilityID() DomainCapabilityID {
	return DomainCapabilityID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewDomainCapabilityIDFromString(value string) (DomainCapabilityID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return DomainCapabilityID{}, err
	}
	return DomainCapabilityID{UUIDValue: uuidValue}, nil
}

func (d DomainCapabilityID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(DomainCapabilityID); ok {
		return d.UUIDValue.EqualsValue(otherID.UUIDValue)
	}
	return false
}
