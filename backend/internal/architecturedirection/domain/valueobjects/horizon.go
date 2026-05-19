package valueobjects

import (
	"errors"

	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrInvalidHorizon = errors.New("horizon must be one of now, next, later")

const (
	HorizonNow   = "now"
	HorizonNext  = "next"
	HorizonLater = "later"
)

type Horizon struct {
	value string
}

func NewHorizon(value string) (Horizon, error) {
	switch value {
	case HorizonNow, HorizonNext, HorizonLater:
		return Horizon{value: value}, nil
	default:
		return Horizon{}, ErrInvalidHorizon
	}
}

func (h Horizon) Value() string { return h.value }

func (h Horizon) Equals(other domain.ValueObject) bool {
	if o, ok := other.(Horizon); ok {
		return h.value == o.value
	}
	return false
}
