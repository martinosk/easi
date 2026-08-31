package valueobjects

import (
	"errors"
	"fmt"

	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrImportanceOutOfRange = errors.New("importance must be between 1 and 5")

var importanceLabels = map[int]string{
	1: "Low",
	2: "Below Average",
	3: "Average",
	4: "Above Average",
	5: "Critical",
}

type Importance struct {
	value int
}

func NewImportance(value int) (Importance, error) {
	if value < 1 || value > 5 {
		return Importance{}, ErrImportanceOutOfRange
	}
	return Importance{value: value}, nil
}

func (i Importance) Value() int {
	return i.value
}

func (i Importance) Label() string {
	return importanceLabels[i.value]
}

func (i Importance) String() string {
	return fmt.Sprintf("%d (%s)", i.value, i.Label())
}

func (i Importance) Equals(other domain.ValueObject) bool {
	if otherImportance, ok := other.(Importance); ok {
		return i.value == otherImportance.value
	}
	return false
}
