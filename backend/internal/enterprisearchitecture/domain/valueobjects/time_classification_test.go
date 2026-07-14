package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTimeClassification_ValidValues(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"tolerate", "Tolerate", "Tolerate"},
		{"invest", "Invest", "Invest"},
		{"migrate", "Migrate", "Migrate"},
		{"eliminate", "Eliminate", "Eliminate"},
		{"lowercase", "invest", "Invest"},
		{"uppercase", "MIGRATE", "Migrate"},
		{"with whitespace", "  Tolerate  ", "Tolerate"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			classification, err := NewTimeClassification(tc.input)

			require.NoError(t, err)
			assert.Equal(t, tc.expected, classification.Value())
		})
	}
}

func TestNewTimeClassification_InvalidValues(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"invalid", "INVALID"},
		{"typo", "TOELRATE"},
		{"partial", "INV"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewTimeClassification(tc.input)

			assert.Error(t, err)
			assert.Equal(t, ErrInvalidTimeClassification, err)
		})
	}
}

func TestTimeClassification_Predicates(t *testing.T) {
	tolerate, _ := NewTimeClassification("TOLERATE")
	invest, _ := NewTimeClassification("INVEST")
	migrate, _ := NewTimeClassification("MIGRATE")
	eliminate, _ := NewTimeClassification("ELIMINATE")

	assert.True(t, tolerate.IsTolerate())
	assert.False(t, tolerate.IsInvest())

	assert.True(t, invest.IsInvest())
	assert.False(t, invest.IsMigrate())

	assert.True(t, migrate.IsMigrate())
	assert.False(t, migrate.IsEliminate())

	assert.True(t, eliminate.IsEliminate())
	assert.False(t, eliminate.IsTolerate())
}

func TestTimeClassification_Equals(t *testing.T) {
	time1, _ := NewTimeClassification("INVEST")
	time2, _ := NewTimeClassification("invest")
	time3, _ := NewTimeClassification("MIGRATE")

	assert.True(t, time1.Equals(time2))
	assert.False(t, time1.Equals(time3))
}

func TestTimeClassification_String(t *testing.T) {
	invest, _ := NewTimeClassification("INVEST")
	assert.Equal(t, "Invest", invest.String())
}
