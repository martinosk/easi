package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createDefaultSections() [4]MaturitySection {
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

	order4, _ := NewSectionOrder(4)
	name4, _ := NewSectionName("Commodity")
	min4, _ := NewMaturityValue(75)
	max4, _ := NewMaturityValue(99)
	section4, _ := NewMaturitySection(order4, name4, min4, max4)

	return [4]MaturitySection{section1, section2, section3, section4}
}

func TestNewMaturityScaleConfig_ValidConfig(t *testing.T) {
	sections := createDefaultSections()

	config, err := NewMaturityScaleConfig(sections)

	require.NoError(t, err)
	assert.Equal(t, 4, len(config.Sections()))
	assert.Equal(t, "Genesis", config.Sections()[0].Name().Value())
	assert.Equal(t, "Commodity", config.Sections()[3].Name().Value())
}

func TestNewMaturityScaleConfig_FirstSectionMustStartAtZero(t *testing.T) {
	order1, _ := NewSectionOrder(1)
	name1, _ := NewSectionName("Genesis")
	min1, _ := NewMaturityValue(5)
	max1, _ := NewMaturityValue(24)
	section1, _ := NewMaturitySection(order1, name1, min1, max1)

	sections := createDefaultSections()
	sections[0] = section1

	_, err := NewMaturityScaleConfig(sections)

	assert.Error(t, err)
	assert.Equal(t, ErrScaleFirstSectionMustStartAtZero, err)
}

func TestNewMaturityScaleConfig_LastSectionMustEndAt99(t *testing.T) {
	order4, _ := NewSectionOrder(4)
	name4, _ := NewSectionName("Commodity")
	min4, _ := NewMaturityValue(75)
	max4, _ := NewMaturityValue(90)
	section4, _ := NewMaturitySection(order4, name4, min4, max4)

	sections := createDefaultSections()
	sections[3] = section4

	_, err := NewMaturityScaleConfig(sections)

	assert.Error(t, err)
	assert.Equal(t, ErrScaleLastSectionMustEndAt99, err)
}

func TestNewMaturityScaleConfig_SectionsMustBeContiguous(t *testing.T) {
	order1, _ := NewSectionOrder(1)
	name1, _ := NewSectionName("Genesis")
	min1, _ := NewMaturityValue(0)
	max1, _ := NewMaturityValue(20)
	section1, _ := NewMaturitySection(order1, name1, min1, max1)

	sections := createDefaultSections()
	sections[0] = section1

	_, err := NewMaturityScaleConfig(sections)

	assert.Error(t, err)
	assert.Equal(t, ErrScaleSectionsMustBeContiguous, err)
}

func TestNewMaturityScaleConfig_SectionsMustHaveCorrectOrder(t *testing.T) {
	order1, _ := NewSectionOrder(2)
	name1, _ := NewSectionName("Genesis")
	min1, _ := NewMaturityValue(0)
	max1, _ := NewMaturityValue(24)
	section1, _ := NewMaturitySection(order1, name1, min1, max1)

	sections := createDefaultSections()
	sections[0] = section1

	_, err := NewMaturityScaleConfig(sections)

	assert.Error(t, err)
	assert.Equal(t, ErrScaleSectionsNotInOrder, err)
}

func TestDefaultMaturityScaleConfig(t *testing.T) {
	config := DefaultMaturityScaleConfig()

	sections := config.Sections()
	assert.Equal(t, 4, len(sections))

	assert.Equal(t, "Genesis", sections[0].Name().Value())
	assert.Equal(t, 0, sections[0].MinValue().Value())
	assert.Equal(t, 24, sections[0].MaxValue().Value())

	assert.Equal(t, "Custom Built", sections[1].Name().Value())
	assert.Equal(t, 25, sections[1].MinValue().Value())
	assert.Equal(t, 49, sections[1].MaxValue().Value())

	assert.Equal(t, "Product", sections[2].Name().Value())
	assert.Equal(t, 50, sections[2].MinValue().Value())
	assert.Equal(t, 74, sections[2].MaxValue().Value())

	assert.Equal(t, "Commodity", sections[3].Name().Value())
	assert.Equal(t, 75, sections[3].MinValue().Value())
	assert.Equal(t, 99, sections[3].MaxValue().Value())
}

func TestMaturityScaleConfig_Equals(t *testing.T) {
	config1 := DefaultMaturityScaleConfig()
	config2 := DefaultMaturityScaleConfig()

	assert.True(t, config1.Equals(config2))
}
