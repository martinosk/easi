package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStrategyPillar_ValidValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected StrategyPillar
	}{
		{"AlwaysOn", "AlwaysOn", PillarAlwaysOn},
		{"Grow", "Grow", PillarGrow},
		{"Transform", "Transform", PillarTransform},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pillar, err := NewStrategyPillar(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, pillar)
		})
	}
}

func TestNewStrategyPillar_TrimSpace(t *testing.T) {
	pillar, err := NewStrategyPillar("  AlwaysOn  ")
	assert.NoError(t, err)
	assert.Equal(t, PillarAlwaysOn, pillar)
}

func TestNewStrategyPillar_Empty(t *testing.T) {
	pillar, err := NewStrategyPillar("")
	assert.NoError(t, err)
	assert.Equal(t, StrategyPillar(""), pillar)
	assert.True(t, pillar.IsEmpty())
}

func TestNewStrategyPillar_InvalidValue(t *testing.T) {
	_, err := NewStrategyPillar("InvalidPillar")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidStrategyPillar, err)
}

func TestStrategyPillar_Value(t *testing.T) {
	pillar := PillarGrow
	assert.Equal(t, "Grow", pillar.Value())
}

func TestStrategyPillar_String(t *testing.T) {
	pillar := PillarTransform
	assert.Equal(t, "Transform", pillar.String())
}

func TestStrategyPillar_Equals(t *testing.T) {
	pillar1 := PillarAlwaysOn
	pillar2 := PillarAlwaysOn
	pillar3 := PillarGrow

	assert.True(t, pillar1.Equals(pillar2))
	assert.False(t, pillar1.Equals(pillar3))
}

func TestStrategyPillar_IsEmpty(t *testing.T) {
	emptyPillar := StrategyPillar("")
	nonEmptyPillar := PillarGrow

	assert.True(t, emptyPillar.IsEmpty())
	assert.False(t, nonEmptyPillar.IsEmpty())
}
