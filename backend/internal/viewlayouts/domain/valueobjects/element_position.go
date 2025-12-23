package valueobjects

import (
	"easi/backend/internal/shared/eventsourcing"
)

type ElementPosition struct {
	elementID   ElementID
	x           float64
	y           float64
	width       *float64
	height      *float64
	customColor *HexColor
	sortOrder   *int
}

func NewElementPosition(elementID ElementID, x, y float64) (ElementPosition, error) {
	return ElementPosition{
		elementID: elementID,
		x:         x,
		y:         y,
	}, nil
}

func NewElementPositionWithOptions(
	elementID ElementID,
	x, y float64,
	width, height *float64,
	customColor *HexColor,
	sortOrder *int,
) (ElementPosition, error) {
	return ElementPosition{
		elementID:   elementID,
		x:           x,
		y:           y,
		width:       width,
		height:      height,
		customColor: customColor,
		sortOrder:   sortOrder,
	}, nil
}

func (e ElementPosition) ElementID() ElementID {
	return e.elementID
}

func (e ElementPosition) X() float64 {
	return e.x
}

func (e ElementPosition) Y() float64 {
	return e.y
}

func (e ElementPosition) Width() *float64 {
	return e.width
}

func (e ElementPosition) Height() *float64 {
	return e.height
}

func (e ElementPosition) CustomColor() *HexColor {
	return e.customColor
}

func (e ElementPosition) SortOrder() *int {
	return e.sortOrder
}

func (e ElementPosition) WithUpdatedPosition(x, y float64) ElementPosition {
	return ElementPosition{
		elementID:   e.elementID,
		x:           x,
		y:           y,
		width:       e.width,
		height:      e.height,
		customColor: e.customColor,
		sortOrder:   e.sortOrder,
	}
}

func (e ElementPosition) Equals(other domain.ValueObject) bool {
	if otherPos, ok := other.(ElementPosition); ok {
		return e.elementID.Equals(otherPos.elementID) &&
			e.x == otherPos.x &&
			e.y == otherPos.y
	}
	return false
}
