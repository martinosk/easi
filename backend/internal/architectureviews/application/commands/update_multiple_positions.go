package commands

type PositionUpdate struct {
	ComponentID string
	X           float64
	Y           float64
}

type UpdateMultiplePositions struct {
	ViewID    string
	Positions []PositionUpdate
}

func (c UpdateMultiplePositions) CommandName() string {
	return "UpdateMultiplePositions"
}
