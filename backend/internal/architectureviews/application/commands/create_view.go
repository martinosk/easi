package commands

// CreateView command
type CreateView struct {
	Name        string
	Description string
	// ID will be populated by the handler after creation
	ID string
}

// CommandName returns the command name
func (c CreateView) CommandName() string {
	return "CreateView"
}
