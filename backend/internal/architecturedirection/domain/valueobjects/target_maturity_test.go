package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTargetMaturity_AcceptsFullRange(t *testing.T) {
	for _, value := range []int{0, 24, 25, 49, 50, 74, 75, 99} {
		m, err := NewTargetMaturity(value)
		require.NoError(t, err)
		assert.Equal(t, value, m.Value())
	}
}

func TestNewTargetMaturity_RejectsOutOfRange(t *testing.T) {
	for _, value := range []int{-1, 100, 250} {
		_, err := NewTargetMaturity(value)
		assert.ErrorIs(t, err, ErrTargetMaturityOutOfRange)
	}
}

func TestTargetMaturity_Sections(t *testing.T) {
	cases := []struct {
		value int
		name  string
		order int
	}{
		{0, "Genesis", 1},
		{24, "Genesis", 1},
		{25, "Custom Build", 2},
		{49, "Custom Build", 2},
		{50, "Product", 3},
		{74, "Product", 3},
		{75, "Commodity", 4},
		{99, "Commodity", 4},
	}
	for _, c := range cases {
		m, err := NewTargetMaturity(c.value)
		require.NoError(t, err)
		assert.Equal(t, c.name, m.SectionName())
		assert.Equal(t, c.order, m.SectionOrder())
	}
}

func TestTargetMaturity_String(t *testing.T) {
	m, err := NewTargetMaturity(65)
	require.NoError(t, err)
	assert.Equal(t, "65 (Product)", m.String())
}

func TestTargetMaturity_Equals(t *testing.T) {
	a, _ := NewTargetMaturity(65)
	b, _ := NewTargetMaturity(65)
	c, _ := NewTargetMaturity(30)
	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
	assert.False(t, a.Equals(JourneyKind{}))
}
