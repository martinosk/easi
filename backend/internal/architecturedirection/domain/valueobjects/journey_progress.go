package valueobjects

import (
	"errors"

	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrInvalidJourneyProgress = errors.New("journey progress must be between 0 and 100")

type JourneyProgress struct {
	value int
}

func NewJourneyProgress(value int) (JourneyProgress, error) {
	if value < 0 || value > 100 {
		return JourneyProgress{}, ErrInvalidJourneyProgress
	}
	return JourneyProgress{value: value}, nil
}

func (p JourneyProgress) Value() int { return p.value }

func (p JourneyProgress) Equals(other domain.ValueObject) bool {
	if o, ok := other.(JourneyProgress); ok {
		return p.value == o.value
	}
	return false
}
