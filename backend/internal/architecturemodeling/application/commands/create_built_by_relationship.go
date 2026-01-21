package commands

type CreateBuiltByRelationship struct {
	InternalTeamID  string
	ComponentID     string
	Notes           string
	ReplaceExisting bool
}

func (c CreateBuiltByRelationship) CommandName() string {
	return "CreateBuiltByRelationship"
}
