package commands

type DeleteApplicationComponent struct {
	ID string
}

func (c DeleteApplicationComponent) CommandName() string {
	return "DeleteApplicationComponent"
}
