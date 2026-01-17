package commands

type DeleteAcquiredEntity struct {
	ID string
}

func (c DeleteAcquiredEntity) CommandName() string {
	return "DeleteAcquiredEntity"
}
