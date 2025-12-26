package valueobjects

import (
	"errors"
	"time"

	"easi/backend/internal/shared/eventsourcing"
)

var (
	ErrTimestampZero = errors.New("timestamp cannot be zero value")
)

type Timestamp struct {
	value time.Time
}

func NewTimestamp(value time.Time) (Timestamp, error) {
	if value.IsZero() {
		return Timestamp{}, ErrTimestampZero
	}
	return Timestamp{value: value.UTC()}, nil
}

func TimestampNow() Timestamp {
	return Timestamp{value: time.Now().UTC()}
}

func (t Timestamp) Value() time.Time {
	return t.value
}

func (t Timestamp) Equals(other domain.ValueObject) bool {
	if otherTs, ok := other.(Timestamp); ok {
		return t.value.Equal(otherTs.value)
	}
	return false
}

func (t Timestamp) String() string {
	return t.value.Format(time.RFC3339)
}
