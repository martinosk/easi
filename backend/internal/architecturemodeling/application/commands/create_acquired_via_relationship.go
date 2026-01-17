package commands

type CreateAcquiredViaRelationship struct {
	AcquiredEntityID string
	ComponentID      string
	Notes            string
}

func (c CreateAcquiredViaRelationship) CommandName() string {
	return "CreateAcquiredViaRelationship"
}
