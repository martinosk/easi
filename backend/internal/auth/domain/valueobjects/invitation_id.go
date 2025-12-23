package valueobjects

import (
	"errors"

	"easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

var ErrInvalidInvitationID = errors.New("invalid invitation ID: must be a valid UUID")

type InvitationID struct {
	sharedvo.UUIDValue
}

func NewInvitationID() InvitationID {
	return InvitationID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewInvitationIDFromString(value string) (InvitationID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return InvitationID{}, ErrInvalidInvitationID
	}
	return InvitationID{UUIDValue: uuidValue}, nil
}

func (id InvitationID) Equals(other domain.ValueObject) bool {
	otherID, ok := other.(InvitationID)
	if !ok {
		return false
	}
	return id.UUIDValue.EqualsValue(otherID.UUIDValue)
}
