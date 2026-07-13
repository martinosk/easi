package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTargetPeriod_Valid_Rule9(t *testing.T) {
	p, err := NewTargetPeriod(2027, 2)
	require.NoError(t, err)
	assert.Equal(t, 2027, p.Year())
	assert.Equal(t, 2, p.Quarter())
}

func TestNewTargetPeriod_YearOutOfRange(t *testing.T) {
	_, err := NewTargetPeriod(1999, 1)
	assert.ErrorIs(t, err, ErrInvalidTargetPeriodYear)

	_, err = NewTargetPeriod(2101, 1)
	assert.ErrorIs(t, err, ErrInvalidTargetPeriodYear)
}

func TestNewTargetPeriod_QuarterOutOfRange(t *testing.T) {
	_, err := NewTargetPeriod(2027, 0)
	assert.ErrorIs(t, err, ErrInvalidTargetPeriodQuarter)

	_, err = NewTargetPeriod(2027, 5)
	assert.ErrorIs(t, err, ErrInvalidTargetPeriodQuarter)
}

func TestTargetPeriod_Before_Rule9(t *testing.T) {
	earlier, _ := NewTargetPeriod(2026, 4)
	later, _ := NewTargetPeriod(2027, 1)
	sameYearEarlierQuarter, _ := NewTargetPeriod(2027, 1)
	sameYearLaterQuarter, _ := NewTargetPeriod(2027, 3)

	assert.True(t, earlier.Before(later))
	assert.False(t, later.Before(earlier))
	assert.True(t, sameYearEarlierQuarter.Before(sameYearLaterQuarter))
	assert.False(t, sameYearLaterQuarter.Before(sameYearEarlierQuarter))
	assert.False(t, later.Before(later))
}

func TestTargetPeriod_Equals(t *testing.T) {
	a, _ := NewTargetPeriod(2027, 2)
	b, _ := NewTargetPeriod(2027, 2)
	c, _ := NewTargetPeriod(2027, 3)
	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
}
