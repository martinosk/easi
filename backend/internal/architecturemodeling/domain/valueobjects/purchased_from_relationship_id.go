package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type PurchasedFromRelationshipID struct {
	sharedvo.UUIDValue
}

func NewPurchasedFromRelationshipID() PurchasedFromRelationshipID {
	return PurchasedFromRelationshipID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewPurchasedFromRelationshipIDFromString(value string) (PurchasedFromRelationshipID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return PurchasedFromRelationshipID{}, err
	}
	return PurchasedFromRelationshipID{UUIDValue: uuidValue}, nil
}

func (p PurchasedFromRelationshipID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(PurchasedFromRelationshipID); ok {
		return p.UUIDValue.EqualsValue(otherID.UUIDValue)
	}
	return false
}
