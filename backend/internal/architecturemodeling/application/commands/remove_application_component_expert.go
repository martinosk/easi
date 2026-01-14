package commands

type RemoveApplicationComponentExpert struct {
	ComponentID string
	ExpertName  string
}

func (c RemoveApplicationComponentExpert) CommandName() string {
	return "RemoveApplicationComponentExpert"
}
