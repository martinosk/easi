package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type ApplicationFitScoreID struct {
	sharedvo.UUIDValue
}

func NewApplicationFitScoreID() ApplicationFitScoreID {
	return ApplicationFitScoreID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewApplicationFitScoreIDFromString(id string) (ApplicationFitScoreID, error) {
	uuid, err := sharedvo.NewUUIDValueFromString(id)
	if err != nil {
		return ApplicationFitScoreID{}, err
	}
	return ApplicationFitScoreID{UUIDValue: uuid}, nil
}

func (a ApplicationFitScoreID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(ApplicationFitScoreID); ok {
		return a.UUIDValue.EqualsValue(otherID.UUIDValue)
	}
	return false
}
