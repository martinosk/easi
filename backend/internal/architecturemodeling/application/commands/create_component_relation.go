package commands

type CreateComponentRelation struct {
	SourceComponentID string
	TargetComponentID string
	RelationType      string
	Name              string
	Description       string
	// ID will be populated by the handler after creation
	ID string
}

func (c CreateComponentRelation) CommandName() string {
	return "CreateComponentRelation"
}
