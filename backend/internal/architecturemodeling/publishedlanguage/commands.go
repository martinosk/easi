package publishedlanguage

type CreateApplicationComponent struct {
	Name        string
	Description string
}

func (c CreateApplicationComponent) CommandName() string {
	return "CreateApplicationComponent"
}

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
