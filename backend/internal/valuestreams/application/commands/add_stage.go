package commands

type AddStage struct {
	ValueStreamID string
	Name          string
	Description   string
	Position      *int
}

func (c AddStage) CommandName() string {
	return "AddStage"
}
