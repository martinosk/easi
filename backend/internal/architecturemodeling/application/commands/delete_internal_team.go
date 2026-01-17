package commands

type DeleteInternalTeam struct {
	ID string
}

func (c DeleteInternalTeam) CommandName() string {
	return "DeleteInternalTeam"
}
