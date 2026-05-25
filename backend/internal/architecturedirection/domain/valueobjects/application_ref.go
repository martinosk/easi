package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type ApplicationRef struct {
	sharedvo.UUIDValue
}

func NewApplicationRef(value string) (ApplicationRef, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return ApplicationRef{}, err
	}
	return ApplicationRef{UUIDValue: uuidValue}, nil
}

func (a ApplicationRef) Equals(other domain.ValueObject) bool {
	if o, ok := other.(ApplicationRef); ok {
		return a.EqualsValue(o.UUIDValue)
	}
	return false
}
