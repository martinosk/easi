package commands

type AddCapabilityExpert struct {
	CapabilityID string
	ExpertName   string
	ExpertRole   string
	ContactInfo  string
}

func (c AddCapabilityExpert) CommandName() string {
	return "AddCapabilityExpert"
}
