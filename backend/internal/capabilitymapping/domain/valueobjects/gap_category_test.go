package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCategorizeGap(t *testing.T) {
	tests := []struct {
		name       string
		gap        int
		importance int
		expected   GapCategory
	}{
		{"liability: gap 2, low importance", 2, 1, GapCategoryLiability},
		{"liability: gap 2, medium importance", 2, 3, GapCategoryLiability},
		{"liability: gap 2, high importance", 2, 5, GapCategoryLiability},
		{"liability: gap 3, any importance", 3, 2, GapCategoryLiability},
		{"liability: gap 4, any importance", 4, 1, GapCategoryLiability},
		{"liability: gap 1, importance 4", 1, 4, GapCategoryLiability},
		{"liability: gap 1, importance 5", 1, 5, GapCategoryLiability},
		{"concern: gap 1, importance 1", 1, 1, GapCategoryConcern},
		{"concern: gap 1, importance 2", 1, 2, GapCategoryConcern},
		{"concern: gap 1, importance 3", 1, 3, GapCategoryConcern},
		{"aligned: gap 0, low importance", 0, 1, GapCategoryAligned},
		{"aligned: gap 0, high importance", 0, 5, GapCategoryAligned},
		{"aligned: negative gap", -1, 3, GapCategoryAligned},
		{"aligned: large negative gap", -3, 5, GapCategoryAligned},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CategorizeGap(tt.gap, tt.importance)
			assert.Equal(t, tt.expected, result)
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
