package commands

// DeleteView command
type DeleteView struct {
	ViewID string
}

// CommandName returns the command name
func (c DeleteView) CommandName() string {
	return "DeleteView"
}
