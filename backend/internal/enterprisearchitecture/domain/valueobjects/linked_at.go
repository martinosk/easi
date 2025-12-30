package valueobjects

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type LinkedAt struct {
	value time.Time
}

func NewLinkedAt() LinkedAt {
	return LinkedAt{value: time.Now().UTC()}
}

func NewLinkedAtFromTime(t time.Time) LinkedAt {
	return LinkedAt{value: t.UTC()}
}

func (l LinkedAt) Value() time.Time {
	return l.value
}

func (l LinkedAt) IsZero() bool {
	return l.value.IsZero()
}

func (l LinkedAt) String() string {
	return l.value.Format(time.RFC3339)
}

func (l LinkedAt) Equals(other domain.ValueObject) bool {
	if otherLA, ok := other.(LinkedAt); ok {
		return l.value.Equal(otherLA.value)
	}
	return false
}
