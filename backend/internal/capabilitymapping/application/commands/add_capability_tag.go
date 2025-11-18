package commands

type AddCapabilityTag struct {
	CapabilityID string
	Tag          string
}

func (c AddCapabilityTag) CommandName() string {
	return "AddCapabilityTag"
}
