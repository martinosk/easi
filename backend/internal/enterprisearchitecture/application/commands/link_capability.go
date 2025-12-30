package commands

type LinkCapability struct {
	EnterpriseCapabilityID string
	DomainCapabilityID     string
	LinkedBy               string
}

func (c LinkCapability) CommandName() string {
	return "LinkCapability"
}
