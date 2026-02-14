package commands

type RemoveDeletedCapability struct {
	CapabilityID string
}

func (c RemoveDeletedCapability) CommandName() string {
	return "RemoveDeletedCapability"
}
