package valueobjects

type GapCategory string

const (
	GapCategoryLiability GapCategory = "liability"
	GapCategoryConcern   GapCategory = "concern"
	GapCategoryAligned   GapCategory = "aligned"
)

const (
	LiabilityGapThreshold   = 2
	ConcernGapThreshold     = 1
	HighImportanceThreshold = 4
)

func CategorizeGap(gap, importance int) GapCategory {
	if gap >= LiabilityGapThreshold || (gap >= ConcernGapThreshold && importance >= HighImportanceThreshold) {
		return GapCategoryLiability
	}
	if gap >= ConcernGapThreshold {
		return GapCategoryConcern
	}
	return GapCategoryAligned
}

func (g GapCategory) String() string {
	return string(g)
}

func (g GapCategory) IsLiability() bool {
	return g == GapCategoryLiability
}

func (g GapCategory) IsConcern() bool {
	return g == GapCategoryConcern
}

func (g GapCategory) IsAligned() bool {
	return g == GapCategoryAligned
}
