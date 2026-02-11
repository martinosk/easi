package commands

type RemoveStageCapability struct {
	ValueStreamID string
	StageID       string
	CapabilityID  string
}

func (c RemoveStageCapability) CommandName() string {
	return "RemoveStageCapability"
}
