package valueobjects

import (
	"errors"
	"fmt"

	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrFitScoreOutOfRange = errors.New("fit score must be between 1 and 5")

var fitScoreLabels = map[int]string{
	1: "Critical",
	2: "Poor",
	3: "Adequate",
	4: "Good",
	5: "Excellent",
}

type FitScore struct {
	value int
}

func NewFitScore(value int) (FitScore, error) {
	if value < 1 || value > 5 {
		return FitScore{}, ErrFitScoreOutOfRange
	}
	return FitScore{value: value}, nil
}

func (f FitScore) Value() int {
	return f.value
}

func (f FitScore) Label() string {
	return fitScoreLabels[f.value]
}

func (f FitScore) String() string {
	return fmt.Sprintf("%d (%s)", f.value, f.Label())
}

func (f FitScore) Equals(other domain.ValueObject) bool {
	if otherScore, ok := other.(FitScore); ok {
		return f.value == otherScore.value
	}
	return false
}
