package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type AcquiredViaRelationshipID struct {
	sharedvo.UUIDValue
}

func NewAcquiredViaRelationshipID() AcquiredViaRelationshipID {
	return AcquiredViaRelationshipID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewAcquiredViaRelationshipIDFromString(value string) (AcquiredViaRelationshipID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return AcquiredViaRelationshipID{}, err
	}
	return AcquiredViaRelationshipID{UUIDValue: uuidValue}, nil
}

func (a AcquiredViaRelationshipID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(AcquiredViaRelationshipID); ok {
		return a.UUIDValue.EqualsValue(otherID.UUIDValue)
	}
	return false
}
