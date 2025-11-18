package commands

type CreateCapabilityDependency struct {
	SourceCapabilityID string
	TargetCapabilityID string
	DependencyType     string
	Description        string
	ID                 string
}

func (c CreateCapabilityDependency) CommandName() string {
	return "CreateCapabilityDependency"
}
