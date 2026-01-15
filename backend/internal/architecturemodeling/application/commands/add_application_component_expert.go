package commands

type AddApplicationComponentExpert struct {
	ComponentID string
	ExpertName  string
	ExpertRole  string
	ContactInfo string
}

func (c AddApplicationComponentExpert) CommandName() string {
	return "AddApplicationComponentExpert"
}
