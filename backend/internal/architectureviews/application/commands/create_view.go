package commands

type CreateView struct {
	Name        string
	Description string
}

// CommandName returns the command name
func (c CreateView) CommandName() string {
	return "CreateView"
}
