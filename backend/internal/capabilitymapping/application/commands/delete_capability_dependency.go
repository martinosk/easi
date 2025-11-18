package commands

type DeleteCapabilityDependency struct {
	ID string
}

func (c DeleteCapabilityDependency) CommandName() string {
	return "DeleteCapabilityDependency"
}
