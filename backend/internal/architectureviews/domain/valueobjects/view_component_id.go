package valueobjects

import (
	"easi/backend/internal/shared/domain"
	sharedvo "easi/backend/internal/shared/domain/valueobjects"
)

type ViewComponentID struct {
	sharedvo.UUIDValue
}

func NewViewComponentID() ViewComponentID {
	return ViewComponentID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewViewComponentIDFromString(value string) (ViewComponentID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return ViewComponentID{}, err
	}
	return ViewComponentID{UUIDValue: uuidValue}, nil
}

func (v ViewComponentID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(ViewComponentID); ok {
		return v.UUIDValue.EqualsValue(otherID.UUIDValue)
	}
	return false
}
