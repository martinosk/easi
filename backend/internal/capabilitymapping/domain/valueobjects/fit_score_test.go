package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFitScore_ValidValues(t *testing.T) {
	tests := []struct {
		value int
		label string
	}{
		{1, "Critical"},
		{2, "Poor"},
		{3, "Adequate"},
		{4, "Good"},
		{5, "Excellent"},
	}

	for _, tt := range tests {
		score, err := NewFitScore(tt.value)
		assert.NoError(t, err)
		assert.Equal(t, tt.value, score.Value())
		assert.Equal(t, tt.label, score.Label())
	}
}

func TestNewFitScore_BelowMinimum(t *testing.T) {
	_, err := NewFitScore(0)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrFitScoreOutOfRange)
}

func TestNewFitScore_AboveMaximum(t *testing.T) {
	_, err := NewFitScore(6)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrFitScoreOutOfRange)
}

func TestNewFitScore_NegativeValue(t *testing.T) {
	_, err := NewFitScore(-1)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrFitScoreOutOfRange)
}

func TestFitScore_String(t *testing.T) {
	score, _ := NewFitScore(3)
	assert.Equal(t, "3 (Adequate)", score.String())
}

func TestFitScore_Equals(t *testing.T) {
	s1, _ := NewFitScore(3)
	s2, _ := NewFitScore(3)
	s3, _ := NewFitScore(4)

	assert.True(t, s1.Equals(s2))
	assert.False(t, s1.Equals(s3))
}

func TestFitScore_Equals_DifferentType(t *testing.T) {
	score, _ := NewFitScore(3)
	rationale, _ := NewFitRationale("test")

	assert.False(t, score.Equals(rationale))
}
