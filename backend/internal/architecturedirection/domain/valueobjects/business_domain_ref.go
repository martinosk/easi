package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type BusinessDomainRef struct {
	sharedvo.UUIDValue
}

func NewBusinessDomainRef(value string) (BusinessDomainRef, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return BusinessDomainRef{}, err
	}
	return BusinessDomainRef{UUIDValue: uuidValue}, nil
}

func (b BusinessDomainRef) Equals(other domain.ValueObject) bool {
	if o, ok := other.(BusinessDomainRef); ok {
		return b.EqualsValue(o.UUIDValue)
	}
	return false
}
