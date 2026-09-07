package valueobjects

import (
	"errors"
	"fmt"

	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrInvalidOwnershipState = errors.New("invalid ownership state: must be unknown, nominated, owned, or managed")

const (
	OwnershipStateUnknown   = "unknown"
	OwnershipStateNominated = "nominated"
	OwnershipStateOwned     = "owned"
	OwnershipStateManaged   = "managed"
)

type OwnershipState struct {
	value string
}

func NewOwnershipState(value string) (OwnershipState, error) {
	switch value {
	case OwnershipStateUnknown, OwnershipStateNominated, OwnershipStateOwned, OwnershipStateManaged:
		return OwnershipState{value: value}, nil
	default:
		return OwnershipState{}, fmt.Errorf("%w: %s", ErrInvalidOwnershipState, value)
	}
}

func UnknownOwnershipState() OwnershipState {
	return OwnershipState{value: OwnershipStateUnknown}
}

func NominatedOwnershipState() OwnershipState {
	return OwnershipState{value: OwnershipStateNominated}
}

func (s OwnershipState) String() string {
	return s.value
}

func (s OwnershipState) IsUnknown() bool {
	return s.value == OwnershipStateUnknown
}

func (s OwnershipState) IsNominated() bool {
	return s.value == OwnershipStateNominated
}

func (s OwnershipState) Equals(other domain.ValueObject) bool {
	if otherState, ok := other.(OwnershipState); ok {
		return s.value == otherState.value
	}
	return false
}
