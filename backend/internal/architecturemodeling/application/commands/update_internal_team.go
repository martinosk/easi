package commands

type UpdateInternalTeam struct {
	ID            string
	Name          string
	Department    string
	ContactPerson string
	Notes         string
}

func (c UpdateInternalTeam) CommandName() string {
	return "UpdateInternalTeam"
}
