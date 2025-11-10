package commands

// RemoveComponentFromView command
type RemoveComponentFromView struct {
	ViewID      string
	ComponentID string
}

// CommandName returns the command name
func (c RemoveComponentFromView) CommandName() string {
	return "RemoveComponentFromView"
}
