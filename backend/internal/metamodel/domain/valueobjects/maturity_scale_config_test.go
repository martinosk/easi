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

type sectionSpec struct {
	order int
	name  string
	min   int
	max   int
}

func buildSection(t *testing.T, spec sectionSpec) MaturitySection {
	t.Helper()
	sectionOrder, err := NewSectionOrder(spec.order)
	require.NoError(t, err)
	sectionName, err := NewSectionName(spec.name)
	require.NoError(t, err)
	minValue, err := NewMaturityValue(spec.min)
	require.NoError(t, err)
	maxValue, err := NewMaturityValue(spec.max)
	require.NoError(t, err)
	section, err := NewMaturitySection(sectionOrder, sectionName, minValue, maxValue)
	require.NoError(t, err)
	return section
}

func TestNewMaturityScaleConfig_InvalidConfig(t *testing.T) {
	tests := []struct {
		name      string
		slot      int
		section   sectionSpec
		wantError error
	}{
		{
			name:      "first section must start at zero",
			slot:      0,
			section:   sectionSpec{order: 1, name: "Genesis", min: 5, max: 24},
			wantError: ErrScaleFirstSectionMustStartAtZero,
		},
		{
			name:      "last section must end at 99",
			slot:      3,
			section:   sectionSpec{order: 4, name: "Commodity", min: 75, max: 90},
			wantError: ErrScaleLastSectionMustEndAt99,
		},
		{
			name:      "sections must be contiguous",
			slot:      0,
			section:   sectionSpec{order: 1, name: "Genesis", min: 0, max: 20},
			wantError: ErrScaleSectionsMustBeContiguous,
		},
		{
			name:      "sections must have correct order",
			slot:      0,
			section:   sectionSpec{order: 2, name: "Genesis", min: 0, max: 24},
			wantError: ErrScaleSectionsNotInOrder,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sections := createDefaultSections()
			sections[tt.slot] = buildSection(t, tt.section)

			_, err := NewMaturityScaleConfig(sections)

			assert.Error(t, err)
			assert.Equal(t, tt.wantError, err)
		})
	}
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
