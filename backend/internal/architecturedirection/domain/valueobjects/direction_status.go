package valueobjects

import (
	"errors"

	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrInvalidDirectionStatus = errors.New("direction status must be one of draft, proposed, agreed, rejected")

const (
	DirectionStatusDraft    = "draft"
	DirectionStatusProposed = "proposed"
	DirectionStatusAgreed   = "agreed"
	DirectionStatusRejected = "rejected"
)

type DirectionStatus struct {
	value string
}

func NewDirectionStatus(value string) (DirectionStatus, error) {
	switch value {
	case DirectionStatusDraft, DirectionStatusProposed, DirectionStatusAgreed, DirectionStatusRejected:
		return DirectionStatus{value: value}, nil
	default:
		return DirectionStatus{}, ErrInvalidDirectionStatus
	}
}

func (s DirectionStatus) Value() string { return s.value }

func (s DirectionStatus) IsActive() bool {
	return s.value != DirectionStatusRejected
}

func (s DirectionStatus) IsTerminal() bool {
	return s.value == DirectionStatusRejected
}

func (s DirectionStatus) IsDraft() bool    { return s.value == DirectionStatusDraft }
func (s DirectionStatus) IsProposed() bool { return s.value == DirectionStatusProposed }
func (s DirectionStatus) IsAgreed() bool   { return s.value == DirectionStatusAgreed }
func (s DirectionStatus) IsRejected() bool { return s.value == DirectionStatusRejected }

func (s DirectionStatus) CanAdvanceTo(target DirectionStatus) bool {
	switch s.value {
	case DirectionStatusDraft:
		return target.value == DirectionStatusProposed
	case DirectionStatusProposed:
		return target.value == DirectionStatusAgreed
	default:
		return false
	}
}

// CanReject allows rejection from any non-terminal state. The BDD scenario
// "At most one active direction per enterprise capability" requires that an
// architect can reject an agreed Direction to capture a new one — that path
// is also the documented escape hatch for the type/horizon/resultingName
// immutability rules (reject-and-replace).
func (s DirectionStatus) CanReject() bool {
	return s.value != DirectionStatusRejected
}

func (s DirectionStatus) Equals(other domain.ValueObject) bool {
	if o, ok := other.(DirectionStatus); ok {
		return s.value == o.value
	}
	return false
}
