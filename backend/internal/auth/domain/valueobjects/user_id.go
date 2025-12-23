package valueobjects

import (
	"errors"

	"easi/backend/internal/shared/domain"
	sharedvo "easi/backend/internal/shared/domain/valueobjects"
)

var ErrInvalidUserID = errors.New("invalid user ID: must be a valid UUID")

type UserID struct {
	sharedvo.UUIDValue
}

func NewUserID() UserID {
	return UserID{UUIDValue: sharedvo.NewUUIDValue()}
}

func UserIDFromString(value string) (UserID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return UserID{}, ErrInvalidUserID
	}
	return UserID{UUIDValue: uuidValue}, nil
}

func (id UserID) Equals(other domain.ValueObject) bool {
	otherID, ok := other.(UserID)
	if !ok {
		return false
	}
	return id.UUIDValue.EqualsValue(otherID.UUIDValue)
}
