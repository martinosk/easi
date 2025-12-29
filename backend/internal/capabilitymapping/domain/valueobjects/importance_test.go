package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewImportance_ValidValues(t *testing.T) {
	tests := []struct {
		value int
		label string
	}{
		{1, "Low"},
		{2, "Below Average"},
		{3, "Average"},
		{4, "Above Average"},
		{5, "Critical"},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			importance, err := NewImportance(tt.value)
			assert.NoError(t, err)
			assert.Equal(t, tt.value, importance.Value())
			assert.Equal(t, tt.label, importance.Label())
		})
	}
}

func TestNewImportance_InvalidValues(t *testing.T) {
	tests := []int{0, -1, 6, 100}

	for _, value := range tests {
		t.Run("value_"+string(rune('0'+value)), func(t *testing.T) {
			_, err := NewImportance(value)
			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrImportanceOutOfRange)
		})
	}
}

func TestImportance_Equals(t *testing.T) {
	imp1, _ := NewImportance(3)
	imp2, _ := NewImportance(3)
	imp3, _ := NewImportance(5)

	assert.True(t, imp1.Equals(imp2))
	assert.False(t, imp1.Equals(imp3))
}

func TestImportance_String(t *testing.T) {
	importance, _ := NewImportance(4)
	assert.Contains(t, importance.String(), "4")
	assert.Contains(t, importance.String(), "Above Average")
}
