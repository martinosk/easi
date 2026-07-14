package valueobjects

import (
	"errors"

	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrInvalidTargetPeriodYear    = errors.New("target period year must be between 2000 and 2100")
	ErrInvalidTargetPeriodQuarter = errors.New("target period quarter must be between 1 and 4")
)

type TargetPeriod struct {
	year    int
	quarter int
}

func NewTargetPeriod(year, quarter int) (TargetPeriod, error) {
	if year < 2000 || year > 2100 {
		return TargetPeriod{}, ErrInvalidTargetPeriodYear
	}
	if quarter < 1 || quarter > 4 {
		return TargetPeriod{}, ErrInvalidTargetPeriodQuarter
	}
	return TargetPeriod{year: year, quarter: quarter}, nil
}

func (p TargetPeriod) Year() int    { return p.year }
func (p TargetPeriod) Quarter() int { return p.quarter }

func (p TargetPeriod) Before(other TargetPeriod) bool {
	if p.year != other.year {
		return p.year < other.year
	}
	return p.quarter < other.quarter
}

func (p TargetPeriod) Equals(other domain.ValueObject) bool {
	if o, ok := other.(TargetPeriod); ok {
		return p.year == o.year && p.quarter == o.quarter
	}
	return false
}
