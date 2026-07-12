package services

import "easi/backend/internal/enterprisearchitecture/domain/valueobjects"

const DefaultGapThreshold = 1.5

type TimeSuggestionResult struct {
	SuggestedTime string
	TechnicalGap  float64
	FunctionalGap float64
	Confidence    string
}

type TimeSuggestionCalculator struct {
	threshold float64
}

func NewTimeSuggestionCalculator(threshold float64) *TimeSuggestionCalculator {
	return &TimeSuggestionCalculator{threshold: threshold}
}

func (c *TimeSuggestionCalculator) Calculate(technicalGaps, functionalGaps []float64) TimeSuggestionResult {
	result := TimeSuggestionResult{}

	hasTechnicalData := len(technicalGaps) > 0
	hasFunctionalData := len(functionalGaps) > 0

	if hasTechnicalData {
		result.TechnicalGap = averageGaps(technicalGaps)
	}
	if hasFunctionalData {
		result.FunctionalGap = averageGaps(functionalGaps)
	}

	result.Confidence = c.determineConfidence(hasTechnicalData, hasFunctionalData, len(technicalGaps), len(functionalGaps))

	if !hasTechnicalData || !hasFunctionalData {
		return result
	}

	result.SuggestedTime = c.determineTimeClassification(result.TechnicalGap, result.FunctionalGap)

	return result
}

func (c *TimeSuggestionCalculator) determineTimeClassification(technicalGap, functionalGap float64) string {
	highTechnicalGap := technicalGap >= c.threshold
	highFunctionalGap := functionalGap >= c.threshold

	if !highTechnicalGap && !highFunctionalGap {
		return valueobjects.TimeClassificationInvest
	}
	if !highTechnicalGap && highFunctionalGap {
		return valueobjects.TimeClassificationTolerate
	}
	if highTechnicalGap && !highFunctionalGap {
		return valueobjects.TimeClassificationMigrate
	}
	return valueobjects.TimeClassificationEliminate
}

func (c *TimeSuggestionCalculator) determineConfidence(hasTechnicalData, hasFunctionalData bool, technicalCount, functionalCount int) string {
	if !hasTechnicalData || !hasFunctionalData {
		return valueobjects.TimeSuggestionConfidenceLow
	}
	if technicalCount >= 2 && functionalCount >= 2 {
		return valueobjects.TimeSuggestionConfidenceHigh
	}
	return valueobjects.TimeSuggestionConfidenceMedium
}

func averageGaps(gaps []float64) float64 {
	if len(gaps) == 0 {
		return 0
	}
	sum := 0.0
	for _, g := range gaps {
		sum += g
	}
	return sum / float64(len(gaps))
}
