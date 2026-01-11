package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCategorizeGap_Liability_LargeGap(t *testing.T) {
	tests := []struct {
		name       string
		gap        int
		importance int
	}{
		{"gap 2, low importance", 2, 1},
		{"gap 2, medium importance", 2, 3},
		{"gap 2, high importance", 2, 5},
		{"gap 3, any importance", 3, 2},
		{"gap 4, any importance", 4, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CategorizeGap(tt.gap, tt.importance)
			assert.Equal(t, GapCategoryLiability, result)
		})
	}
}

func TestCategorizeGap_Liability_HighImportanceWithSmallGap(t *testing.T) {
	tests := []struct {
		name       string
		gap        int
		importance int
	}{
		{"gap 1, importance 4", 1, 4},
		{"gap 1, importance 5", 1, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CategorizeGap(tt.gap, tt.importance)
			assert.Equal(t, GapCategoryLiability, result)
		})
	}
}

func TestCategorizeGap_Concern(t *testing.T) {
	tests := []struct {
		name       string
		gap        int
		importance int
	}{
		{"gap 1, importance 1", 1, 1},
		{"gap 1, importance 2", 1, 2},
		{"gap 1, importance 3", 1, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CategorizeGap(tt.gap, tt.importance)
			assert.Equal(t, GapCategoryConcern, result)
		})
	}
}

func TestCategorizeGap_Aligned(t *testing.T) {
	tests := []struct {
		name       string
		gap        int
		importance int
	}{
		{"gap 0, low importance", 0, 1},
		{"gap 0, high importance", 0, 5},
		{"negative gap", -1, 3},
		{"large negative gap", -3, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CategorizeGap(tt.gap, tt.importance)
			assert.Equal(t, GapCategoryAligned, result)
		})
	}
}

func TestCategorizeGap_BoundaryConditions(t *testing.T) {
	assert.Equal(t, GapCategoryAligned, CategorizeGap(0, 5), "gap 0 should be aligned regardless of importance")
	assert.Equal(t, GapCategoryConcern, CategorizeGap(1, 3), "gap 1 with importance 3 should be concern")
	assert.Equal(t, GapCategoryLiability, CategorizeGap(1, 4), "gap 1 with importance 4 should be liability")
	assert.Equal(t, GapCategoryLiability, CategorizeGap(2, 1), "gap 2 should be liability regardless of importance")
}

func TestGapCategory_String(t *testing.T) {
	assert.Equal(t, "liability", GapCategoryLiability.String())
	assert.Equal(t, "concern", GapCategoryConcern.String())
	assert.Equal(t, "aligned", GapCategoryAligned.String())
}

func TestGapCategory_IsLiability(t *testing.T) {
	assert.True(t, GapCategoryLiability.IsLiability())
	assert.False(t, GapCategoryConcern.IsLiability())
	assert.False(t, GapCategoryAligned.IsLiability())
}

func TestGapCategory_IsConcern(t *testing.T) {
	assert.False(t, GapCategoryLiability.IsConcern())
	assert.True(t, GapCategoryConcern.IsConcern())
	assert.False(t, GapCategoryAligned.IsConcern())
}

func TestGapCategory_IsAligned(t *testing.T) {
	assert.False(t, GapCategoryLiability.IsAligned())
	assert.False(t, GapCategoryConcern.IsAligned())
	assert.True(t, GapCategoryAligned.IsAligned())
}

func TestGapThresholdConstants(t *testing.T) {
	assert.Equal(t, 2, LiabilityGapThreshold)
	assert.Equal(t, 1, ConcernGapThreshold)
	assert.Equal(t, 4, HighImportanceThreshold)
}
