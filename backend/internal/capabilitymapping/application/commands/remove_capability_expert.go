package commands

type RemoveCapabilityExpert struct {
	CapabilityID string
	ExpertName   string
	ExpertRole   string
	ContactInfo  string
}

func (c RemoveCapabilityExpert) CommandName() string {
	return "RemoveCapabilityExpert"
}
