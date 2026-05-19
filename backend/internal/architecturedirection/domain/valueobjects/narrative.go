package valueobjects

import (
	"errors"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
)

const MaxNarrativeLength = 1000

var ErrNarrativeTooLong = errors.New("narrative cannot exceed 1000 characters")

type Narrative struct {
	value string
}

func NewNarrative(value string) (Narrative, error) {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) > MaxNarrativeLength {
		return Narrative{}, ErrNarrativeTooLong
	}
	return Narrative{value: trimmed}, nil
}

func EmptyNarrative() Narrative { return Narrative{} }

func (n Narrative) Value() string  { return n.value }
func (n Narrative) IsEmpty() bool  { return n.value == "" }
func (n Narrative) String() string { return n.value }

func (n Narrative) Equals(other domain.ValueObject) bool {
	if o, ok := other.(Narrative); ok {
		return n.value == o.value
	}
	return false
}
