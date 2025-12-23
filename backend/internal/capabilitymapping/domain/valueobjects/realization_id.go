package valueobjects

import (
	"easi/backend/internal/shared/domain"
	sharedvo "easi/backend/internal/shared/domain/valueobjects"
)

type RealizationID struct {
	sharedvo.UUIDValue
}

func NewRealizationID() RealizationID {
	return RealizationID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewRealizationIDFromString(value string) (RealizationID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return RealizationID{}, err
	}
	return RealizationID{UUIDValue: uuidValue}, nil
}

func (r RealizationID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(RealizationID); ok {
		return r.UUIDValue.EqualsValue(otherID.UUIDValue)
	}
	return false
}
