package valueobjects

import (
	"errors"

	"easi/backend/internal/shared/domain"
)

var ErrInvalidInvitationStatus = errors.New("invalid invitation status: must be pending, accepted, expired, or revoked")

type InvitationStatus struct {
	value string
}

var (
	InvitationStatusPending  = InvitationStatus{value: "pending"}
	InvitationStatusAccepted = InvitationStatus{value: "accepted"}
	InvitationStatusExpired  = InvitationStatus{value: "expired"}
	InvitationStatusRevoked  = InvitationStatus{value: "revoked"}
)

func InvitationStatusFromString(s string) (InvitationStatus, error) {
	switch s {
	case "pending":
		return InvitationStatusPending, nil
	case "accepted":
		return InvitationStatusAccepted, nil
	case "expired":
		return InvitationStatusExpired, nil
	case "revoked":
		return InvitationStatusRevoked, nil
	default:
		return InvitationStatus{}, ErrInvalidInvitationStatus
	}
}

func (s InvitationStatus) String() string {
	return s.value
}

func (s InvitationStatus) IsPending() bool {
	return s.value == "pending"
}

func (s InvitationStatus) IsAccepted() bool {
	return s.value == "accepted"
}

func (s InvitationStatus) IsExpired() bool {
	return s.value == "expired"
}

func (s InvitationStatus) IsRevoked() bool {
	return s.value == "revoked"
}

func (s InvitationStatus) CanTransitionTo(target InvitationStatus) bool {
	if !s.IsPending() {
		return false
	}
	return target.IsAccepted() || target.IsExpired() || target.IsRevoked()
}

func (s InvitationStatus) Equals(other domain.ValueObject) bool {
	otherStatus, ok := other.(InvitationStatus)
	if !ok {
		return false
	}
	return s.value == otherStatus.value
}
