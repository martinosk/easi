package commands

type DeleteEnterpriseCapability struct {
	ID string
}

func (c DeleteEnterpriseCapability) CommandName() string {
	return "DeleteEnterpriseCapability"
}
