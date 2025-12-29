package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type PillarID struct {
	sharedvo.UUIDValue
}

func NewPillarID() PillarID {
	return PillarID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewPillarIDFromString(value string) (PillarID, error) {
	uuid, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return PillarID{}, err
	}
	return PillarID{UUIDValue: uuid}, nil
}

func (p PillarID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(PillarID); ok {
		return p.UUIDValue.EqualsValue(otherID.UUIDValue)
	}
	return false
}
