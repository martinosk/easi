package valueobjects

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type SetAt struct {
	value time.Time
}

func NewSetAt() SetAt {
	return SetAt{value: time.Now().UTC()}
}

func NewSetAtFromTime(t time.Time) SetAt {
	return SetAt{value: t.UTC()}
}

func (s SetAt) Value() time.Time {
	return s.value
}

func (s SetAt) IsZero() bool {
	return s.value.IsZero()
}

func (s SetAt) String() string {
	return s.value.Format(time.RFC3339)
}

func (s SetAt) Equals(other domain.ValueObject) bool {
	if otherSA, ok := other.(SetAt); ok {
		return s.value.Equal(otherSA.value)
	}
	return false
}
