package commands

// CreateApplicationComponent command
type CreateApplicationComponent struct {
	Name        string
	Description string
	// ID will be populated by the handler after creation
	ID string
}

// CommandName returns the command name
func (c CreateApplicationComponent) CommandName() string {
	return "CreateApplicationComponent"
}
