package commands

type DeleteBuiltByRelationship struct {
	ID string
}

func (c DeleteBuiltByRelationship) CommandName() string {
	return "DeleteBuiltByRelationship"
}
