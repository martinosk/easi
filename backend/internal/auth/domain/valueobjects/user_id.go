package valueobjects

import (
	"errors"

	"github.com/google/uuid"

	"easi/backend/internal/shared/domain"
)

var ErrInvalidUserID = errors.New("invalid user ID: must be a valid UUID")

type UserID struct {
	value string
}

func NewUserID() UserID {
	return UserID{value: uuid.New().String()}
}

func UserIDFromString(value string) (UserID, error) {
	if value == "" {
		return UserID{}, ErrInvalidUserID
	}
	if _, err := uuid.Parse(value); err != nil {
		return UserID{}, ErrInvalidUserID
	}
	return UserID{value: value}, nil
}

func (id UserID) Value() string {
	return id.value
}

func (id UserID) String() string {
	return id.value
}

func (id UserID) Equals(other domain.ValueObject) bool {
	otherID, ok := other.(UserID)
	if !ok {
		return false
	}
	return id.value == otherID.value
}
