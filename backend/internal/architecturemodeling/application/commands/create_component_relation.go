package commands

type CreateComponentRelation struct {
	SourceComponentID string
	TargetComponentID string
	RelationType      string
	Name              string
	Description       string
}

func (c CreateComponentRelation) CommandName() string {
	return "CreateComponentRelation"
}
