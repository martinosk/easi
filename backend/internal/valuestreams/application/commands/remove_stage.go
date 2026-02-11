package commands

type RemoveStage struct {
	ValueStreamID string
	StageID       string
}

func (c RemoveStage) CommandName() string {
	return "RemoveStage"
}
