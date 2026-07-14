package valueobjects

import (
	"errors"

	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrInvalidJourneyStatus = errors.New("journey status must be one of planned, in-flight, done, abandoned")

const (
	JourneyStatusPlanned   = "planned"
	JourneyStatusInFlight  = "in-flight"
	JourneyStatusDone      = "done"
	JourneyStatusAbandoned = "abandoned"
)

type JourneyStatus struct {
	value string
}

func NewJourneyStatus(value string) (JourneyStatus, error) {
	switch value {
	case JourneyStatusPlanned, JourneyStatusInFlight, JourneyStatusDone, JourneyStatusAbandoned:
		return JourneyStatus{value: value}, nil
	default:
		return JourneyStatus{}, ErrInvalidJourneyStatus
	}
}

func (s JourneyStatus) Value() string { return s.value }

func (s JourneyStatus) IsActive() bool {
	return s.value == JourneyStatusPlanned || s.value == JourneyStatusInFlight
}

func (s JourneyStatus) IsTerminal() bool { return !s.IsActive() }

func (s JourneyStatus) CanStart() bool    { return s.value == JourneyStatusPlanned }
func (s JourneyStatus) CanComplete() bool { return s.value == JourneyStatusInFlight }
func (s JourneyStatus) CanAbandon() bool  { return s.IsActive() }

func (s JourneyStatus) Equals(other domain.ValueObject) bool {
	if o, ok := other.(JourneyStatus); ok {
		return s.value == o.value
	}
	return false
}
