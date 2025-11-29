package commands

type AssignCapabilityToDomain struct {
	BusinessDomainID string
	CapabilityID     string
	AssignmentID     string
}

func (c AssignCapabilityToDomain) CommandName() string {
	return "AssignCapabilityToDomain"
}
