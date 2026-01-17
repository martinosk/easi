package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type BuiltByRelationshipID struct {
	sharedvo.UUIDValue
}

func NewBuiltByRelationshipID() BuiltByRelationshipID {
	return BuiltByRelationshipID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewBuiltByRelationshipIDFromString(value string) (BuiltByRelationshipID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return BuiltByRelationshipID{}, err
	}
	return BuiltByRelationshipID{UUIDValue: uuidValue}, nil
}

func (b BuiltByRelationshipID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(BuiltByRelationshipID); ok {
		return b.UUIDValue.EqualsValue(otherID.UUIDValue)
	}
	return false
}
