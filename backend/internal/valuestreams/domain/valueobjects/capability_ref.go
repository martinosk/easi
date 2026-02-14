package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type CapabilityRef struct {
	sharedvo.UUIDValue
}

func NewCapabilityRef(value string) (CapabilityRef, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return CapabilityRef{}, err
	}
	return CapabilityRef{UUIDValue: uuidValue}, nil
}

func MustNewCapabilityRef(value string) CapabilityRef {
	ref, err := NewCapabilityRef(value)
	if err != nil {
		panic(err)
	}
	return ref
}

func (c CapabilityRef) Equals(other domain.ValueObject) bool {
	if otherRef, ok := other.(CapabilityRef); ok {
		return c.UUIDValue.EqualsValue(otherRef.UUIDValue)
	}
	return false
}
