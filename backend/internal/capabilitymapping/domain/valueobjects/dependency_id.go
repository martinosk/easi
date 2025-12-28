package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type DependencyID struct {
	sharedvo.UUIDValue
}

func NewDependencyID() DependencyID {
	return DependencyID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewDependencyIDFromString(value string) (DependencyID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return DependencyID{}, err
	}
	return DependencyID{UUIDValue: uuidValue}, nil
}

func (d DependencyID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(DependencyID); ok {
		return d.UUIDValue.EqualsValue(otherID.UUIDValue)
	}
	return false
}
