package commands

type StagePositionEntry struct {
	StageID  string
	Position int
}

type ReorderStages struct {
	ValueStreamID string
	Positions     []StagePositionEntry
}

func (c ReorderStages) CommandName() string {
	return "ReorderStages"
}
