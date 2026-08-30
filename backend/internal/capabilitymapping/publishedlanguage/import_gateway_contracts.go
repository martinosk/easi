package publishedlanguage

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
