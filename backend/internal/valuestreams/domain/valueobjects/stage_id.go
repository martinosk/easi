package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type StageID struct {
	sharedvo.UUIDValue
}

func NewStageID() StageID {
	return StageID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewStageIDFromString(value string) (StageID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return StageID{}, err
	}
	return StageID{UUIDValue: uuidValue}, nil
}

func (v StageID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(StageID); ok {
		return v.UUIDValue.EqualsValue(otherID.UUIDValue)
	}
	return false
}
