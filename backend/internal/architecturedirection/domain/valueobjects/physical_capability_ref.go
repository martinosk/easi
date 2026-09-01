package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type PhysicalCapabilityRef struct {
	sharedvo.UUIDValue
}

func NewPhysicalCapabilityRef(value string) (PhysicalCapabilityRef, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return PhysicalCapabilityRef{}, err
	}
	return PhysicalCapabilityRef{UUIDValue: uuidValue}, nil
}

func (p PhysicalCapabilityRef) Equals(other domain.ValueObject) bool {
	if o, ok := other.(PhysicalCapabilityRef); ok {
		return p.EqualsValue(o.UUIDValue)
	}
	return false
}
