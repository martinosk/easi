package publishedlanguage

type ImportedOption struct {
	ID     string
	Label  string
	Active bool
}

type ImportedAttribute struct {
	ID       string
	Name     string
	Type     string
	HelpText string
	Options  []ImportedOption
	Min      *float64
	Max      *float64
	Active   bool
}

type ImportSubjectAttribute struct {
	TenantID    string
	SubjectType string
	Attribute   ImportedAttribute
	ImportedBy  string
}

func (c ImportSubjectAttribute) CommandName() string {
	return "ImportSubjectAttribute"
}
