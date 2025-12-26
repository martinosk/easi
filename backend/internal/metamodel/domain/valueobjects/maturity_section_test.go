package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMaturitySection_ValidSection(t *testing.T) {
	order, _ := NewSectionOrder(1)
	name, _ := NewSectionName("Genesis")
	minValue, _ := NewMaturityValue(0)
	maxValue, _ := NewMaturityValue(24)

	section, err := NewMaturitySection(order, name, minValue, maxValue)

	require.NoError(t, err)
	assert.Equal(t, 1, section.Order().Value())
	assert.Equal(t, "Genesis", section.Name().Value())
	assert.Equal(t, 0, section.MinValue().Value())
	assert.Equal(t, 24, section.MaxValue().Value())
}

func TestNewMaturitySection_MinGreaterThanMax(t *testing.T) {
	order, _ := NewSectionOrder(1)
	name, _ := NewSectionName("Genesis")
	minValue, _ := NewMaturityValue(50)
	maxValue, _ := NewMaturityValue(24)

	_, err := NewMaturitySection(order, name, minValue, maxValue)

	assert.Error(t, err)
	assert.Equal(t, ErrMaturitySectionInvalidRange, err)
}

func TestMaturitySection_Equals(t *testing.T) {
	order, _ := NewSectionOrder(1)
	name, _ := NewSectionName("Genesis")
	minValue, _ := NewMaturityValue(0)
	maxValue, _ := NewMaturityValue(24)

	section1, _ := NewMaturitySection(order, name, minValue, maxValue)
	section2, _ := NewMaturitySection(order, name, minValue, maxValue)

	differentName, _ := NewSectionName("Different")
	section3, _ := NewMaturitySection(order, differentName, minValue, maxValue)

	assert.True(t, section1.Equals(section2))
	assert.False(t, section1.Equals(section3))
}

func TestMaturitySection_AdjacentTo(t *testing.T) {
	order1, _ := NewSectionOrder(1)
	name1, _ := NewSectionName("Genesis")
	min1, _ := NewMaturityValue(0)
	max1, _ := NewMaturityValue(24)
	section1, _ := NewMaturitySection(order1, name1, min1, max1)

	order2, _ := NewSectionOrder(2)
	name2, _ := NewSectionName("Custom Built")
	min2, _ := NewMaturityValue(25)
	max2, _ := NewMaturityValue(49)
	section2, _ := NewMaturitySection(order2, name2, min2, max2)

	order3, _ := NewSectionOrder(3)
	name3, _ := NewSectionName("Product")
	min3, _ := NewMaturityValue(50)
	max3, _ := NewMaturityValue(74)
	section3, _ := NewMaturitySection(order3, name3, min3, max3)

	assert.True(t, section1.IsAdjacentTo(section2))
	assert.False(t, section1.IsAdjacentTo(section3))
}
