package commands

type AddStageCapability struct {
	ValueStreamID string
	StageID       string
	CapabilityID  string
}

func (c AddStageCapability) CommandName() string {
	return "AddStageCapability"
}
