package valueobjects

import (
	"easi/backend/internal/shared/domain"
	sharedvo "easi/backend/internal/shared/domain/valueobjects"
)

type LayoutContainerID struct {
	sharedvo.UUIDValue
}

func NewLayoutContainerID() LayoutContainerID {
	return LayoutContainerID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewLayoutContainerIDFromString(value string) (LayoutContainerID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return LayoutContainerID{}, err
	}
	return LayoutContainerID{UUIDValue: uuidValue}, nil
}

func (l LayoutContainerID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(LayoutContainerID); ok {
		return l.UUIDValue.EqualsValue(otherID.UUIDValue)
	}
	return false
}
