package commands

type UnassignCapabilityFromDomain struct {
	AssignmentID string
}

func (c UnassignCapabilityFromDomain) CommandName() string {
	return "UnassignCapabilityFromDomain"
}
