package valueobjects

import (
	"errors"

	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

var (
	ErrInvalidAttributeID = errors.New("attribute ID must be a valid UUID")
	ErrInvalidOptionID    = errors.New("option ID must be a valid UUID")
)

type AttributeID struct {
	sharedvo.UUIDValue
}

func NewAttributeID() AttributeID {
	return AttributeID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewAttributeIDFromString(value string) (AttributeID, error) {
	id, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return AttributeID{}, ErrInvalidAttributeID
	}
	return AttributeID{UUIDValue: id}, nil
}

func (a AttributeID) Equals(other domain.ValueObject) bool {
	if o, ok := other.(AttributeID); ok {
		return a.EqualsValue(o.UUIDValue)
	}
	return false
}

type OptionID struct {
	sharedvo.UUIDValue
}

func NewOptionID() OptionID {
	return OptionID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewOptionIDFromString(value string) (OptionID, error) {
	id, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return OptionID{}, ErrInvalidOptionID
	}
	return OptionID{UUIDValue: id}, nil
}

func (o OptionID) Equals(other domain.ValueObject) bool {
	if other2, ok := other.(OptionID); ok {
		return o.EqualsValue(other2.UUIDValue)
	}
	return false
}
