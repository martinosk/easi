package commands

type LinkSystemToCapability struct {
	CapabilityID     string
	ComponentID      string
	RealizationLevel string
	Notes            string
}

func (c LinkSystemToCapability) CommandName() string {
	return "LinkSystemToCapability"
}
