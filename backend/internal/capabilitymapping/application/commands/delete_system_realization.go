package commands

type DeleteSystemRealization struct {
	ID string
}

func (c DeleteSystemRealization) CommandName() string {
	return "DeleteSystemRealization"
}
