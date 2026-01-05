package valueobjects

import (
	"errors"

	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrOwnerUserIDEmpty = errors.New("owner user ID cannot be empty")

type ViewOwner struct {
	userID string
	email  string
}

func NewViewOwner(userID, email string) (ViewOwner, error) {
	if userID == "" {
		return ViewOwner{}, ErrOwnerUserIDEmpty
	}
	return ViewOwner{userID: userID, email: email}, nil
}

func EmptyViewOwner() ViewOwner {
	return ViewOwner{}
}

func (v ViewOwner) UserID() string {
	return v.userID
}

func (v ViewOwner) Email() string {
	return v.email
}

func (v ViewOwner) IsEmpty() bool {
	return v.userID == ""
}

func (v ViewOwner) Equals(other domain.ValueObject) bool {
	if otherOwner, ok := other.(ViewOwner); ok {
		return v.userID == otherOwner.userID
	}
	return false
}
