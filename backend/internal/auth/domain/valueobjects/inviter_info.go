package valueobjects

import (
	"errors"

	"github.com/google/uuid"

	"easi/backend/internal/shared/eventsourcing"
)

var (
	ErrInvalidInviterID    = errors.New("inviter ID cannot be empty")
	ErrInvalidInviterEmail = errors.New("inviter email cannot be empty")
)

type InviterInfo struct {
	userID uuid.UUID
	email  Email
}

func NewInviterInfo(userID uuid.UUID, email Email) (InviterInfo, error) {
	if userID == uuid.Nil {
		return InviterInfo{}, ErrInvalidInviterID
	}
	if email.Value() == "" {
		return InviterInfo{}, ErrInvalidInviterEmail
	}
	return InviterInfo{
		userID: userID,
		email:  email,
	}, nil
}

func (i InviterInfo) UserID() uuid.UUID {
	return i.userID
}

func (i InviterInfo) Email() Email {
	return i.email
}

func (i InviterInfo) Equals(other domain.ValueObject) bool {
	otherInfo, ok := other.(InviterInfo)
	if !ok {
		return false
	}
	return i.userID == otherInfo.userID && i.email.Equals(otherInfo.email)
}
