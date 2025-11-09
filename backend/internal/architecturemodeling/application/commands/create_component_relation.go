package commands

// CreateComponentRelation command
type CreateComponentRelation struct {
	SourceComponentID string
	TargetComponentID string
	RelationType      string
	Name              string
	Description       string
	// ID will be populated by the handler after creation
	ID string
}

// CommandName returns the command name
func (c CreateComponentRelation) CommandName() string {
	return "CreateComponentRelation"
}
