package commands

// UpdateComponentPosition command
type UpdateComponentPosition struct {
	ViewID      string
	ComponentID string
	X           float64
	Y           float64
}

// CommandName returns the command name
func (c UpdateComponentPosition) CommandName() string {
	return "UpdateComponentPosition"
}
