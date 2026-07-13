package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type CapabilityJourneyID struct {
	sharedvo.UUIDValue
}

func NewCapabilityJourneyID() CapabilityJourneyID {
	return CapabilityJourneyID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewCapabilityJourneyIDFromString(value string) (CapabilityJourneyID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return CapabilityJourneyID{}, err
	}
	return CapabilityJourneyID{UUIDValue: uuidValue}, nil
}

func (i CapabilityJourneyID) Equals(other domain.ValueObject) bool {
	if o, ok := other.(CapabilityJourneyID); ok {
		return i.EqualsValue(o.UUIDValue)
	}
	return false
}
