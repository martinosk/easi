package publishedlanguage

type CreateValueStream struct {
	Name        string
	Description string
}

func (c CreateValueStream) CommandName() string {
	return "CreateValueStream"
}

type AddStage struct {
	ValueStreamID string
	Name          string
	Description   string
	Position      *int
}

func (c AddStage) CommandName() string {
	return "AddStage"
}

type AddStageCapability struct {
	ValueStreamID string
	StageID       string
	CapabilityID  string
}

func (c AddStageCapability) CommandName() string {
	return "AddStageCapability"
}
