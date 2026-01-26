package valueobjects

import (
	"errors"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
)

const (
	TimeClassificationTolerate  = "TOLERATE"
	TimeClassificationInvest    = "INVEST"
	TimeClassificationMigrate   = "MIGRATE"
	TimeClassificationEliminate = "ELIMINATE"
)

var ErrInvalidTimeClassification = errors.New("time classification must be TOLERATE, INVEST, MIGRATE, or ELIMINATE")

type TimeClassification struct {
	value string
}

func NewTimeClassification(value string) (TimeClassification, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	if !isValidTimeClassification(normalized) {
		return TimeClassification{}, ErrInvalidTimeClassification
	}
	return TimeClassification{value: normalized}, nil
}

func isValidTimeClassification(value string) bool {
	return value == TimeClassificationTolerate ||
		value == TimeClassificationInvest ||
		value == TimeClassificationMigrate ||
		value == TimeClassificationEliminate
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
