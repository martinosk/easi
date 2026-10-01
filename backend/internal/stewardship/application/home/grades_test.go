package home

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeSharesFromCounts_Distribution(t *testing.T) {
	shares := timeSharesOf(GradeCounts{Invest: 20, Tolerate: 8, Migrate: 4, Eliminate: 2, NotAssessed: 6})

	assert.Equal(t, &TimeShares{Invest: 50, Tolerate: 20, Migrate: 10, Eliminate: 5, NotAssessed: 15}, shares)
}

func TestTimeSharesFromCounts_SumToExactlyHundredByLargestRemainder(t *testing.T) {
	shares := timeSharesOf(GradeCounts{Invest: 1, Tolerate: 1, Migrate: 1})

	assert.Equal(t, &TimeShares{Invest: 34, Tolerate: 33, Migrate: 33}, shares)
}

func TestTimeSharesFromCounts_LargestRemainderGetsTheExtraPoint(t *testing.T) {
	shares := timeSharesOf(GradeCounts{Invest: 1, Eliminate: 2})

	assert.Equal(t, &TimeShares{Invest: 33, Eliminate: 67}, shares)
	assert.Equal(t, 100, shares.Invest+shares.Tolerate+shares.Migrate+shares.Eliminate+shares.NotAssessed)
}

func TestTimeSharesFromCounts_NoRealisationsHasNoShares(t *testing.T) {
	assert.Nil(t, timeSharesOf(GradeCounts{}))
}

func TestDominantGrade(t *testing.T) {
	tests := []struct {
		name   string
		counts GradeCounts
		want   string
	}{
		{"most frequent wins", GradeCounts{Invest: 2, Eliminate: 1}, GradeInvest},
		{"invest-tolerate tie", GradeCounts{Invest: 1, Tolerate: 1}, GradeTolerate},
		{"invest-migrate tie", GradeCounts{Invest: 1, Migrate: 1}, GradeMigrate},
		{"invest-eliminate tie", GradeCounts{Invest: 1, Eliminate: 1}, GradeEliminate},
		{"tolerate-migrate tie", GradeCounts{Tolerate: 1, Migrate: 1}, GradeMigrate},
		{"tolerate-eliminate tie", GradeCounts{Tolerate: 1, Eliminate: 1}, GradeEliminate},
		{"migrate-eliminate tie", GradeCounts{Migrate: 2, Eliminate: 2}, GradeEliminate},
		{"three-way tie", GradeCounts{Invest: 1, Tolerate: 1, Migrate: 1}, GradeMigrate},
		{"not assessed ignored", GradeCounts{Invest: 1, NotAssessed: 5}, GradeInvest},
		{"nothing assessed", GradeCounts{NotAssessed: 3}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.counts.dominant())
		})
	}
}

func TestGradeCounts_AddMapsRecordedGrades(t *testing.T) {
	var counts GradeCounts
	require.NoError(t, counts.Add("Invest", 2))
	require.NoError(t, counts.Add("Eliminate", 1))
	require.NoError(t, counts.Add("", 4))

	assert.Equal(t, GradeCounts{Invest: 2, Eliminate: 1, NotAssessed: 4}, counts)
	assert.Equal(t, 7, counts.total())
}

func TestGradeCounts_AddRefusesAGradeItDoesNotKnow(t *testing.T) {
	var counts GradeCounts

	err := counts.Add("Retire", 3)

	assert.ErrorIs(t, err, ErrUnknownGrade)
	assert.Zero(t, counts.total())
}
