package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type StrategyImportanceID struct {
	sharedvo.UUIDValue
}

func NewStrategyImportanceID() StrategyImportanceID {
	return StrategyImportanceID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewStrategyImportanceIDFromString(value string) (StrategyImportanceID, error) {
	uuid, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return StrategyImportanceID{}, err
	}
	return StrategyImportanceID{UUIDValue: uuid}, nil
}

func (s StrategyImportanceID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(StrategyImportanceID); ok {
		return s.UUIDValue.EqualsValue(otherID.UUIDValue)
	}
	return false
}
