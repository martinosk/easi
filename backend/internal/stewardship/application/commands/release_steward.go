package commands

type ReleaseSteward struct {
	DomainID   string
	Concern    string
	ReleasedBy string
}

func (c ReleaseSteward) CommandName() string {
	return "ReleaseSteward"
}
