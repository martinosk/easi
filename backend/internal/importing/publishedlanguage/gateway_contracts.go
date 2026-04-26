package publishedlanguage

type CreateRelationInput struct {
	SourceID     string
	TargetID     string
	RelationType string
	Name         string
	Description  string
}

type CreateCapabilityInput struct {
	Name        string
	Description string
	ParentID    string
	Level       string
}

type LinkSystemInput struct {
	CapabilityID     string
	ComponentID      string
	RealizationLevel string
	Notes            string
}
