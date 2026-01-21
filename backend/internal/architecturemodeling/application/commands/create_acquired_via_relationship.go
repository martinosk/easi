package commands

type CreateAcquiredViaRelationship struct {
	AcquiredEntityID string
	ComponentID      string
	Notes            string
	ReplaceExisting  bool
}

func (c CreateAcquiredViaRelationship) CommandName() string {
	return "CreateAcquiredViaRelationship"
}
