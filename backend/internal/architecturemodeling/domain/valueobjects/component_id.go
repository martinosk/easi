package valueobjects

import (
	"easi/backend/internal/shared/domain"
	sharedvo "easi/backend/internal/shared/domain/valueobjects"
)

type ComponentID struct {
	sharedvo.UUIDValue
}

func NewComponentID() ComponentID {
	return ComponentID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewComponentIDFromString(value string) (ComponentID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return ComponentID{}, err
	}
	return ComponentID{UUIDValue: uuidValue}, nil
}

func (c ComponentID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(ComponentID); ok {
		return c.UUIDValue.EqualsValue(otherID.UUIDValue)
	}
	return false
}
