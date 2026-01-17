package commands

type CreateBuiltByRelationship struct {
	InternalTeamID string
	ComponentID    string
	Notes          string
}

func (c CreateBuiltByRelationship) CommandName() string {
	return "CreateBuiltByRelationship"
}
