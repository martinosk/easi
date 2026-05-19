package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type EnterpriseCapabilityRef struct {
	sharedvo.UUIDValue
}

func NewEnterpriseCapabilityRef(value string) (EnterpriseCapabilityRef, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return EnterpriseCapabilityRef{}, err
	}
	return EnterpriseCapabilityRef{UUIDValue: uuidValue}, nil
}

func (e EnterpriseCapabilityRef) Equals(other domain.ValueObject) bool {
	if o, ok := other.(EnterpriseCapabilityRef); ok {
		return e.EqualsValue(o.UUIDValue)
	}
	return false
}

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
