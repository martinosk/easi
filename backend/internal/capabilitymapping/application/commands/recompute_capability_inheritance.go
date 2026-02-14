package commands

type RecomputeCapabilityInheritance struct {
	CapabilityID string
}

func (c RecomputeCapabilityInheritance) CommandName() string {
	return "RecomputeCapabilityInheritance"
}
