package commands

// SetDefaultView command
type SetDefaultView struct {
	ViewID string
}

// CommandName returns the command name
func (c SetDefaultView) CommandName() string {
	return "SetDefaultView"
}
