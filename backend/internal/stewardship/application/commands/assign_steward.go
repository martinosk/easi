package commands

type AssignSteward struct {
	DomainID   string
	Concern    string
	StewardID  string
	AssignedBy string
}

func (c AssignSteward) CommandName() string {
	return "AssignSteward"
}
