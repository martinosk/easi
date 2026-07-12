package valueobjects

import (
	"errors"

	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrInvalidTimeGrade = errors.New("time grade must be one of Invest, Tolerate, Migrate, Eliminate")

const (
	TimeGradeInvest    = "Invest"
	TimeGradeTolerate  = "Tolerate"
	TimeGradeMigrate   = "Migrate"
	TimeGradeEliminate = "Eliminate"
)

type TimeGrade struct {
	value string
}

func NewTimeGrade(value string) (TimeGrade, error) {
	switch value {
	case TimeGradeInvest, TimeGradeTolerate, TimeGradeMigrate, TimeGradeEliminate:
		return TimeGrade{value: value}, nil
	default:
		return TimeGrade{}, ErrInvalidTimeGrade
	}
}

func (g TimeGrade) Value() string { return g.value }

func (g TimeGrade) Equals(other domain.ValueObject) bool {
	if otherGrade, ok := other.(TimeGrade); ok {
		return g.value == otherGrade.value
	}
	return false
}
