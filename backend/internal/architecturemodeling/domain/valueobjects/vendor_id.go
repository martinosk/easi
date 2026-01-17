package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type VendorID struct {
	sharedvo.UUIDValue
}

func NewVendorID() VendorID {
	return VendorID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewVendorIDFromString(value string) (VendorID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return VendorID{}, err
	}
	return VendorID{UUIDValue: uuidValue}, nil
}

func (v VendorID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(VendorID); ok {
		return v.UUIDValue.EqualsValue(otherID.UUIDValue)
	}
	return false
}
