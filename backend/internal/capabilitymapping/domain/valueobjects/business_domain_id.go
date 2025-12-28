package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type BusinessDomainID struct {
	sharedvo.UUIDValue
}

func NewBusinessDomainID() BusinessDomainID {
	return BusinessDomainID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewBusinessDomainIDFromString(value string) (BusinessDomainID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return BusinessDomainID{}, err
	}
	return BusinessDomainID{UUIDValue: uuidValue}, nil
}

func (b BusinessDomainID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(BusinessDomainID); ok {
		return b.UUIDValue.EqualsValue(otherID.UUIDValue)
	}
	return false
}
