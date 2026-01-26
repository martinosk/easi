package valueobjects

import (
	"errors"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
)

const (
	TimeSuggestionConfidenceLow    = "LOW"
	TimeSuggestionConfidenceMedium = "MEDIUM"
	TimeSuggestionConfidenceHigh   = "HIGH"
)

var ErrInvalidTimeSuggestionConfidence = errors.New("time suggestion confidence must be LOW, MEDIUM, or HIGH")

type TimeSuggestionConfidence struct {
	value string
}

func NewTimeSuggestionConfidence(value string) (TimeSuggestionConfidence, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	if !isValidTimeSuggestionConfidence(normalized) {
		return TimeSuggestionConfidence{}, ErrInvalidTimeSuggestionConfidence
	}
	return TimeSuggestionConfidence{value: normalized}, nil
}

func isValidTimeSuggestionConfidence(value string) bool {
	return value == TimeSuggestionConfidenceLow ||
		value == TimeSuggestionConfidenceMedium ||
		value == TimeSuggestionConfidenceHigh
}

func (c TimeSuggestionConfidence) Value() string {
	return c.value
}

func (c TimeSuggestionConfidence) IsLow() bool {
	return c.value == TimeSuggestionConfidenceLow
}

func (c TimeSuggestionConfidence) IsMedium() bool {
	return c.value == TimeSuggestionConfidenceMedium
}

func (c TimeSuggestionConfidence) IsHigh() bool {
	return c.value == TimeSuggestionConfidenceHigh
}

func (c TimeSuggestionConfidence) Equals(other domain.ValueObject) bool {
	if otherConf, ok := other.(TimeSuggestionConfidence); ok {
		return c.value == otherConf.value
	}
	return false
}

func (c TimeSuggestionConfidence) String() string {
	return c.value
}
