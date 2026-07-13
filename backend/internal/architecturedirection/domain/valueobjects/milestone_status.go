package valueobjects

import (
	"errors"

	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrInvalidMilestoneStatus = errors.New("milestone status must be one of planned, in-flight, done")

const (
	MilestoneStatusPlanned  = "planned"
	MilestoneStatusInFlight = "in-flight"
	MilestoneStatusDone     = "done"
)

type MilestoneStatus struct {
	value string
}

func NewMilestoneStatus(value string) (MilestoneStatus, error) {
	switch value {
	case MilestoneStatusPlanned, MilestoneStatusInFlight, MilestoneStatusDone:
		return MilestoneStatus{value: value}, nil
	default:
		return MilestoneStatus{}, ErrInvalidMilestoneStatus
	}
}

func (s MilestoneStatus) Value() string { return s.value }

func (s MilestoneStatus) Equals(other domain.ValueObject) bool {
	if o, ok := other.(MilestoneStatus); ok {
		return s.value == o.value
	}
	return false
}
