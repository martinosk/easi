package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateTimeSuggestion_AllQuadrants(t *testing.T) {
	testCases := []struct {
		name               string
		technicalGaps      []float64
		functionalGaps     []float64
		expectedTime       string
		expectedConfidence string
	}{
		{
			name:               "INVEST - low tech gap, low func gap",
			technicalGaps:      []float64{0.5, 1.0, 0.8},
			functionalGaps:     []float64{0.3, 0.7, 1.2},
			expectedTime:       "INVEST",
			expectedConfidence: "HIGH",
		},
		{
			name:               "TOLERATE - low tech gap, high func gap",
			technicalGaps:      []float64{0.5, 1.0},
			functionalGaps:     []float64{2.0, 1.8},
			expectedTime:       "TOLERATE",
			expectedConfidence: "HIGH",
		},
		{
			name:               "MIGRATE - high tech gap, low func gap",
			technicalGaps:      []float64{2.5, 1.8},
			functionalGaps:     []float64{0.5, 0.8},
			expectedTime:       "MIGRATE",
			expectedConfidence: "HIGH",
		},
		{
			name:               "ELIMINATE - high tech gap, high func gap",
			technicalGaps:      []float64{2.0, 2.5},
			functionalGaps:     []float64{1.8, 2.2},
			expectedTime:       "ELIMINATE",
			expectedConfidence: "HIGH",
		},
	}

	calculator := NewTimeSuggestionCalculator(DefaultGapThreshold)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calculator.Calculate(tc.technicalGaps, tc.functionalGaps)

			assert.Equal(t, tc.expectedTime, result.SuggestedTime)
			assert.Equal(t, tc.expectedConfidence, result.Confidence)
		})
	}
}

func TestCalculateTimeSuggestion_ConfidenceLevels(t *testing.T) {
	calculator := NewTimeSuggestionCalculator(DefaultGapThreshold)

	t.Run("HIGH confidence - all pillars scored", func(t *testing.T) {
		result := calculator.Calculate(
			[]float64{1.0, 0.8, 1.2},
			[]float64{0.5, 0.7, 0.9},
		)
		assert.Equal(t, "HIGH", result.Confidence)
	})

	t.Run("MEDIUM confidence - at least one of each type", func(t *testing.T) {
		result := calculator.Calculate(
			[]float64{1.0},
			[]float64{0.5},
		)
		assert.Equal(t, "MEDIUM", result.Confidence)
	})

	t.Run("LOW confidence - only tech gaps", func(t *testing.T) {
		result := calculator.Calculate(
			[]float64{1.0, 1.2},
			[]float64{},
		)
		assert.Equal(t, "LOW", result.Confidence)
		assert.Empty(t, result.SuggestedTime)
	})

	t.Run("LOW confidence - only func gaps", func(t *testing.T) {
		result := calculator.Calculate(
			[]float64{},
			[]float64{0.5, 0.7},
		)
		assert.Equal(t, "LOW", result.Confidence)
		assert.Empty(t, result.SuggestedTime)
	})

	t.Run("LOW confidence - no data", func(t *testing.T) {
		result := calculator.Calculate(
			[]float64{},
			[]float64{},
		)
		assert.Equal(t, "LOW", result.Confidence)
		assert.Empty(t, result.SuggestedTime)
	})
}

func TestCalculateTimeSuggestion_GapAverages(t *testing.T) {
	calculator := NewTimeSuggestionCalculator(DefaultGapThreshold)

	result := calculator.Calculate(
		[]float64{1.0, 2.0, 3.0},
		[]float64{0.5, 1.5},
	)

	assert.InDelta(t, 2.0, result.TechnicalGap, 0.01)
	assert.InDelta(t, 1.0, result.FunctionalGap, 0.01)
}

func TestCalculateTimeSuggestion_ThresholdBoundary(t *testing.T) {
	calculator := NewTimeSuggestionCalculator(1.5)

	t.Run("exactly at threshold is HIGH gap", func(t *testing.T) {
		result := calculator.Calculate(
			[]float64{1.5},
			[]float64{1.0},
		)
		assert.Equal(t, "MIGRATE", result.SuggestedTime)
	})

	t.Run("just below threshold is LOW gap", func(t *testing.T) {
		result := calculator.Calculate(
			[]float64{1.49},
			[]float64{1.0},
		)
		assert.Equal(t, "INVEST", result.SuggestedTime)
	})
}

func TestCalculateTimeSuggestion_NegativeGaps(t *testing.T) {
	calculator := NewTimeSuggestionCalculator(DefaultGapThreshold)

	result := calculator.Calculate(
		[]float64{-1.0, 0.5},
		[]float64{-0.5, 0.3},
	)

	assert.Equal(t, "INVEST", result.SuggestedTime)
	assert.InDelta(t, -0.25, result.TechnicalGap, 0.01)
	assert.InDelta(t, -0.1, result.FunctionalGap, 0.01)
}
