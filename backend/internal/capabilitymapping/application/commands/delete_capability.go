package commands

type DeleteCapability struct {
	ID string
}

func (c DeleteCapability) CommandName() string {
	return "DeleteCapability"
}
