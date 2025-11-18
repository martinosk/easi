package commands

type DeleteComponentRelation struct {
	ID string
}

func (c DeleteComponentRelation) CommandName() string {
	return "DeleteComponentRelation"
}
