package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPillarWeight_ValidValues(t *testing.T) {
	tests := []struct {
		name  string
		value int
	}{
		{"Zero", 0},
		{"Minimum", 1},
		{"Middle", 50},
		{"Maximum", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weight, err := NewPillarWeight(tt.value)
			assert.NoError(t, err)
			assert.Equal(t, tt.value, weight.Value())
		})
	}
}

func TestNewPillarWeight_NegativeValue(t *testing.T) {
	_, err := NewPillarWeight(-1)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidPillarWeight, err)
}

func TestNewPillarWeight_OverMaximum(t *testing.T) {
	_, err := NewPillarWeight(101)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidPillarWeight, err)
}

func TestPillarWeight_Value(t *testing.T) {
	weight, _ := NewPillarWeight(75)
	assert.Equal(t, 75, weight.Value())
}

func TestPillarWeight_Equals(t *testing.T) {
	weight1, _ := NewPillarWeight(50)
	weight2, _ := NewPillarWeight(50)
	weight3, _ := NewPillarWeight(75)

	assert.True(t, weight1.Equals(weight2))
	assert.False(t, weight1.Equals(weight3))
}
