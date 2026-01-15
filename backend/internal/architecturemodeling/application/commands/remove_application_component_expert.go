package commands

type RemoveApplicationComponentExpert struct {
	ComponentID string
	ExpertName  string
	ExpertRole  string
	ContactInfo string
}

func (c RemoveApplicationComponentExpert) CommandName() string {
	return "RemoveApplicationComponentExpert"
}
