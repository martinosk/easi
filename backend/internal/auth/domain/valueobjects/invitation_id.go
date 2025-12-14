package valueobjects

import (
	"errors"

	"github.com/google/uuid"

	"easi/backend/internal/shared/domain"
)

var ErrInvalidInvitationID = errors.New("invalid invitation ID: must be a valid UUID")

type InvitationID struct {
	value string
}

func NewInvitationID() InvitationID {
	return InvitationID{value: uuid.New().String()}
}

func NewInvitationIDFromString(value string) (InvitationID, error) {
	if value == "" {
		return InvitationID{}, ErrInvalidInvitationID
	}
	if _, err := uuid.Parse(value); err != nil {
		return InvitationID{}, ErrInvalidInvitationID
	}
	return InvitationID{value: value}, nil
}

func (id InvitationID) Value() string {
	return id.value
}

func (id InvitationID) String() string {
	return id.value
}

func (id InvitationID) Equals(other domain.ValueObject) bool {
	otherID, ok := other.(InvitationID)
	if !ok {
		return false
	}
	return id.value == otherID.value
}
