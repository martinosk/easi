package commands

type CreateInternalTeam struct {
	Name          string
	Department    string
	ContactPerson string
	Notes         string
}

func (c CreateInternalTeam) CommandName() string {
	return "CreateInternalTeam"
}
