package commands

type UpdateStage struct {
	ValueStreamID string
	StageID       string
	Name          string
	Description   string
}

func (c UpdateStage) CommandName() string {
	return "UpdateStage"
}
