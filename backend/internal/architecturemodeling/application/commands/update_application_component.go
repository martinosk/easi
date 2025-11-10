package commands

// UpdateApplicationComponent command
type UpdateApplicationComponent struct {
	ID          string
	Name        string
	Description string
}

// CommandName returns the command name
func (c UpdateApplicationComponent) CommandName() string {
	return "UpdateApplicationComponent"
}
