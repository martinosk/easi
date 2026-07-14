package valueobjects

import (
	"errors"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
)

const (
	TimeClassificationTolerate  = "Tolerate"
	TimeClassificationInvest    = "Invest"
	TimeClassificationMigrate   = "Migrate"
	TimeClassificationEliminate = "Eliminate"
)

var ErrInvalidTimeClassification = errors.New("time classification must be Tolerate, Invest, Migrate, or Eliminate")

type TimeClassification struct {
	value string
}

func NewTimeClassification(value string) (TimeClassification, error) {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "TOLERATE":
		return TimeClassification{value: TimeClassificationTolerate}, nil
	case "INVEST":
		return TimeClassification{value: TimeClassificationInvest}, nil
	case "MIGRATE":
		return TimeClassification{value: TimeClassificationMigrate}, nil
	case "ELIMINATE":
		return TimeClassification{value: TimeClassificationEliminate}, nil
	}
	return TimeClassification{}, ErrInvalidTimeClassification
}

func (t TimeClassification) Value() string {
	return t.value
}

func (t TimeClassification) IsTolerate() bool {
	return t.value == TimeClassificationTolerate
}

func (t TimeClassification) IsInvest() bool {
	return t.value == TimeClassificationInvest
}

func (t TimeClassification) IsMigrate() bool {
	return t.value == TimeClassificationMigrate
}

func (t TimeClassification) IsEliminate() bool {
	return t.value == TimeClassificationEliminate
}

func (t TimeClassification) Equals(other domain.ValueObject) bool {
	if otherTime, ok := other.(TimeClassification); ok {
		return t.value == otherTime.value
	}
	return false
}

func (t TimeClassification) String() string {
	return t.value
}
