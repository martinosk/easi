package commands

type DeleteAcquiredViaRelationship struct {
	ID string
}

func (c DeleteAcquiredViaRelationship) CommandName() string {
	return "DeleteAcquiredViaRelationship"
}
