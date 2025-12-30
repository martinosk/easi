package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewImportance_ValidValues(t *testing.T) {
	testCases := []struct {
		value         int
		expectedLabel string
	}{
		{1, "Low"},
		{2, "Below Average"},
		{3, "Average"},
		{4, "Above Average"},
		{5, "Critical"},
	}

	for _, tc := range testCases {
		t.Run(tc.expectedLabel, func(t *testing.T) {
			importance, err := NewImportance(tc.value)
			require.NoError(t, err)
			assert.Equal(t, tc.value, importance.Value())
			assert.Equal(t, tc.expectedLabel, importance.Label())
		})
	}
}

func TestNewImportance_RejectsZero(t *testing.T) {
	_, err := NewImportance(0)
	assert.ErrorIs(t, err, ErrImportanceOutOfRange)
}

func TestNewImportance_RejectsSix(t *testing.T) {
	_, err := NewImportance(6)
	assert.ErrorIs(t, err, ErrImportanceOutOfRange)
}

func TestNewImportance_RejectsNegativeValues(t *testing.T) {
	_, err := NewImportance(-1)
	assert.ErrorIs(t, err, ErrImportanceOutOfRange)

	_, err = NewImportance(-100)
	assert.ErrorIs(t, err, ErrImportanceOutOfRange)
}

func TestImportance_String(t *testing.T) {
	importance, _ := NewImportance(3)
	assert.Equal(t, "3 (Average)", importance.String())
}

func TestImportance_Equals(t *testing.T) {
	imp1, _ := NewImportance(3)
	imp2, _ := NewImportance(3)
	imp3, _ := NewImportance(4)

	assert.True(t, imp1.Equals(imp2))
	assert.False(t, imp1.Equals(imp3))
}
