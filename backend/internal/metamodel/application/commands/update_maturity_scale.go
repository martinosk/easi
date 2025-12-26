package commands

type MaturitySectionInput struct {
	Order    int
	Name     string
	MinValue int
	MaxValue int
}

type UpdateMaturityScale struct {
	ID         string
	Sections   [4]MaturitySectionInput
	ModifiedBy string
}

func (c UpdateMaturityScale) CommandName() string {
	return "UpdateMaturityScale"
}
