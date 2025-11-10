package commands

// RenameView command
type RenameView struct {
	ViewID  string
	NewName string
}

// CommandName returns the command name
func (c RenameView) CommandName() string {
	return "RenameView"
}
