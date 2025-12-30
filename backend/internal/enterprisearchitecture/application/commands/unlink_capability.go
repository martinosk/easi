package commands

type UnlinkCapability struct {
	LinkID string
}

func (c UnlinkCapability) CommandName() string {
	return "UnlinkCapability"
}
