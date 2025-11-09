package commands

// AddComponentToView command
type AddComponentToView struct {
	ViewID      string
	ComponentID string
	X           float64
	Y           float64
}

// CommandName returns the command name
func (c AddComponentToView) CommandName() string {
	return "AddComponentToView"
}
