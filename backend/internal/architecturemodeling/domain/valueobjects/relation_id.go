package valueobjects

import (
	"easi/backend/internal/shared/domain"
	sharedvo "easi/backend/internal/shared/domain/valueobjects"
)

type RelationID struct {
	sharedvo.UUIDValue
}

func NewRelationID() RelationID {
	return RelationID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewRelationIDFromString(value string) (RelationID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return RelationID{}, err
	}
	return RelationID{UUIDValue: uuidValue}, nil
}

func (r RelationID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(RelationID); ok {
		return r.UUIDValue.EqualsValue(otherID.UUIDValue)
	}
	return false
}
