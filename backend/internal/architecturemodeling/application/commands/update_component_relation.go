package commands

// UpdateComponentRelation command
type UpdateComponentRelation struct {
	ID          string
	Name        string
	Description string
}

// CommandName returns the command name
func (c UpdateComponentRelation) CommandName() string {
	return "UpdateComponentRelation"
}
