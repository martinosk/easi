package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type StrategyPillarID struct {
	sharedvo.UUIDValue
}

func NewStrategyPillarID() StrategyPillarID {
	return StrategyPillarID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewStrategyPillarIDFromString(value string) (StrategyPillarID, error) {
	uuid, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return StrategyPillarID{}, err
	}
	return StrategyPillarID{UUIDValue: uuid}, nil
}

func (s StrategyPillarID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(StrategyPillarID); ok {
		return s.UUIDValue.EqualsValue(otherID.UUIDValue)
	}
	return false
}
