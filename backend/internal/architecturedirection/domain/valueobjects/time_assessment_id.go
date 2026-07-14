package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type TimeAssessmentID struct {
	sharedvo.UUIDValue
}

func NewTimeAssessmentID() TimeAssessmentID {
	return TimeAssessmentID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewTimeAssessmentIDFromString(value string) (TimeAssessmentID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return TimeAssessmentID{}, err
	}
	return TimeAssessmentID{UUIDValue: uuidValue}, nil
}

func (i TimeAssessmentID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(TimeAssessmentID); ok {
		return i.EqualsValue(otherID.UUIDValue)
	}
	return false
}
