package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
)

type ViewVisibility struct {
	isPrivate bool
}

var (
	VisibilityPrivate = ViewVisibility{isPrivate: true}
	VisibilityPublic  = ViewVisibility{isPrivate: false}
)

func NewViewVisibility(isPrivate bool) ViewVisibility {
	return ViewVisibility{isPrivate: isPrivate}
}

func (v ViewVisibility) IsPrivate() bool {
	return v.isPrivate
}

func (v ViewVisibility) IsPublic() bool {
	return !v.isPrivate
}

func (v ViewVisibility) Equals(other domain.ValueObject) bool {
	if otherVis, ok := other.(ViewVisibility); ok {
		return v.isPrivate == otherVis.isPrivate
	}
	return false
}
